package export

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	protocol_pb "github.com/HonLaderDev/HonLader-core-api/pb/minecraft/protocol"
	packet_pb "github.com/HonLaderDev/HonLader-core-api/pb/minecraft/protocol/packet"
	hdefine "github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils/structure"
	"github.com/HonLaderDev/bedrock-world-operator/chunk"
	bwo_define "github.com/HonLaderDev/bedrock-world-operator/define"
	mc_packet "github.com/HonLaderDev/mousetunnel/minecraft/protocol/packet"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

type exportBatch struct {
	pos       hdefine.ChunkPos
	subChunks []hdefine.SubChunkPos
}

func (e *ExportTask) run(ctx context.Context) error {
	defer e.finishTaskContext(ctx)
	defer func() {
		if e.world != nil {
			_ = e.world.CloseWorld()
		}
	}()
	defer func() {
		if e.tempDir != "" {
			_ = os.RemoveAll(e.tempDir)
		}
	}()

	if err := e.checkRunContext(ctx); err != nil {
		return err
	}

	batches := e.buildBatches()
	e.publish(EventNameRunStart, e.regionSize(), len(batches))

	for i := e.CurrentBatch; i < len(batches); i++ {
		if err := e.checkRunContext(ctx); err != nil {
			return err
		}
		batch := batches[i]
		e.publish(EventNameRunBatchStart, i)

		if err := e.moveToBatch(ctx, batch.pos); err != nil {
			return fmt.Errorf("ExportTask.run: move to batch: %w", err)
		}
		e.publish(EventNameRunBatchMove, batch.pos)

		if err := e.exportBatch(ctx, batch); err != nil {
			return fmt.Errorf("ExportTask.run: export batch: %w", err)
		}
		e.CurrentBatch = i + 1
		e.publish(EventNameRunBatchFinish)
	}

	if err := e.finishWorld(); err != nil {
		return err
	}
	if err := e.packWorld(); err != nil {
		return err
	}
	e.publish(EventNameRunFinish, e.outputPath)
	return nil
}

func (e *ExportTask) checkRunContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (e *ExportTask) regionSize() hdefine.Size {
	start, end := e.StartPos.SortStartAndEndPos(e.EndPos)
	return hdefine.BlockSize(start, end)
}

func (e *ExportTask) buildBatches() []exportBatch {
	start, end := e.StartPos.SortStartAndEndPos(e.EndPos)
	startChunkX := int32(start.X() >> 4)
	endChunkX := int32(end.X() >> 4)
	startChunkZ := int32(start.Z() >> 4)
	endChunkZ := int32(end.Z() >> 4)
	startSubChunkY := int32(start.Y() >> 4)
	endSubChunkY := int32(end.Y() >> 4)

	side := e.ChunkBatchSide
	if side <= 0 {
		side = defaultChunkBatchSide
	}

	var batches []exportBatch
	for x := startChunkX; x <= endChunkX; x += int32(side) {
		for z := startChunkZ; z <= endChunkZ; z += int32(side) {
			batchEndX := minInt32(x+int32(side)-1, endChunkX)
			batchEndZ := minInt32(z+int32(side)-1, endChunkZ)
			var subChunks []hdefine.SubChunkPos
			for cx := x; cx <= batchEndX; cx++ {
				for cz := z; cz <= batchEndZ; cz++ {
					for sy := startSubChunkY; sy <= endSubChunkY; sy++ {
						subChunks = append(subChunks, hdefine.SubChunkPos{cx, sy, cz})
					}
				}
			}
			batches = append(batches, exportBatch{
				pos:       hdefine.ChunkPos{x, z},
				subChunks: subChunks,
			})
		}
	}
	return batches
}

func (e *ExportTask) moveToBatch(ctx context.Context, pos hdefine.ChunkPos) error {
	_ = ctx
	_ = pos
	return nil
}

func (e *ExportTask) exportBatch(ctx context.Context, batch exportBatch) error {
	if len(batch.subChunks) == 0 {
		return nil
	}
	expected := len(batch.subChunks)
	var mu sync.Mutex
	done := make(chan struct{})
	received := 0
	listenerID, err := e.frame.Client().Resources().PacketListener().ListenPacket(
		ctx,
		[]uint32{mc_packet.IDSubChunk},
		func(pk *packet_pb.Packet, err error) {
			if err != nil {
				e.publish(EventNameRunSubChunkDecodeFailed, hdefine.SubChunkPos{}, err)
				return
			}
			if pk == nil || pk.GetValue() == nil {
				return
			}
			resp := new(packet_pb.SubChunk)
			if err := pk.GetValue().UnmarshalTo(resp); err != nil {
				e.publish(EventNameRunSubChunkDecodeFailed, hdefine.SubChunkPos{}, err)
				return
			}
			if err := e.handleSubChunkPacket(resp); err != nil {
				e.publish(EventNameRunSubChunkDecodeFailed, hdefine.SubChunkPos{}, err)
				return
			}
			mu.Lock()
			received += len(resp.GetSubChunkEntries())
			if received >= expected {
				select {
				case <-done:
				default:
					close(done)
				}
			}
			mu.Unlock()
		},
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = e.frame.Client().Resources().PacketListener().DestroyListener(context.Background(), listenerID)
	}()

	req := &packet_pb.SubChunkRequest{
		Dimension: int32(e.Dimension),
		Position:  &protocol_pb.SubChunkPos{Value: []int32{batch.pos.X(), 0, batch.pos.Z()}},
		Offsets:   make([]*protocol_pb.SubChunkOffset, 0, len(batch.subChunks)),
	}
	for _, pos := range batch.subChunks {
		offset := protocol_pb.SubChunkOffset{Value: []int32{pos.X() - batch.pos.X(), pos.Y(), pos.Z() - batch.pos.Z()}}
		req.Offsets = append(req.Offsets, &offset)
	}
	anyValue, err := anypb.New(req)
	if err != nil {
		return err
	}
	e.publish(EventNameRunSubChunkRequest, len(req.Offsets))
	if err := e.frame.Client().Resources().WritePacket(ctx, &packet_pb.Packet{Value: anyValue}); err != nil {
		return err
	}
	select {
	case <-done:
	case <-time.After(e.RequestTimeout):
		e.publish(EventNameRunSubChunkTimeout, expected)
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (e *ExportTask) handleSubChunkPacket(resp *packet_pb.SubChunk) error {
	if resp == nil || resp.GetPosition() == nil {
		return nil
	}
	center := resp.GetPosition()
	centerValue := center.GetValue()
	if len(centerValue) < 3 {
		return nil
	}
	subChunkStart := hdefine.SubChunkPos{centerValue[0], centerValue[1], centerValue[2]}
	for _, entry := range resp.GetSubChunkEntries() {
		if entry == nil || entry.GetOffset() == nil {
			continue
		}
		values := entry.GetOffset().GetValue()
		if len(values) < 3 {
			continue
		}
		pos := hdefine.SubChunkPos{
			subChunkStart.X() + values[0],
			subChunkStart.Y() + values[1],
			subChunkStart.Z() + values[2],
		}
		switch entry.GetResult() {
		case 1, 6:
			sub, _, err := chunk.DecodeSubChunk(bytes.NewBuffer(entry.GetRawPayload()), bwo_define.Dimension(e.Dimension).Range(), chunk.NetworkEncoding, e.world.BlockRuntimeIDTable())
			if err != nil {
				return err
			}
			if err := e.world.SaveSubChunk(e.Dimension, pos, sub); err != nil {
				e.publish(EventNameRunSubChunkSaveFailed, pos, err)
				continue
			}
		case 2:
			continue
		}
	}
	return nil
}

func (e *ExportTask) finishWorld() error {
	if e.world == nil {
		return nil
	}
	worldName := fmt.Sprintf("%s@[%d,%d,%d]~[%d,%d,%d]", filepath.Base(e.FilePath), e.StartPos.X(), e.StartPos.Y(), e.StartPos.Z(), e.EndPos.X(), e.EndPos.Y(), e.EndPos.Z())
	e.world.LevelDat().LevelName = worldName
	if err := e.world.UpdateLevelDat(); err != nil {
		return fmt.Errorf("ExportTask.finishWorld: update level dat: %w", err)
	}
	if err := e.world.CloseWorld(); err != nil {
		return fmt.Errorf("ExportTask.finishWorld: close world: %w", err)
	}
	return nil
}

func (e *ExportTask) packWorld() error {
	if e.tempDir == "" {
		return nil
	}
	output := e.FilePath
	if !strings.EqualFold(filepath.Ext(output), ".mcworld") {
		output += ".mcworld"
	}
	if !filepath.IsAbs(output) {
		output = filepath.Clean(output)
	}
	e.outputPath = output
	return structure.ZipDirectoryContents(e.worldDir, output)
}

func minInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
