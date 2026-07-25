package build

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/HonLaderDev/HonLader-core-api/pb/minecraft/protocol"
	packet_pb "github.com/HonLaderDev/HonLader-core-api/pb/minecraft/protocol/packet"
	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils"
	"github.com/HonLaderDev/bedrock-world-operator/chunk"
	"google.golang.org/protobuf/types/known/anypb"
)

const temporaryNBTStructureName = "HonLader_NBT_block"

// buildChunkGroupNBT 按配置写入当前区块组内的命令方块和其他 NBT 方块。
func (b *BuildTask) buildChunkGroupNBT(ctx context.Context, chunks map[define.ChunkPos]*chunk.Chunk, nbts map[define.ChunkPos][]map[string]any) error {
	if len(nbts) == 0 || (b.IgnoreCommandBlock && b.IgnoreOtherNBTBlock) {
		return nil
	}

	total := b.countChunkGroupNBT(chunks, nbts)
	if total == 0 {
		return nil
	}
	b.publish(EventNameRunNBTStart, total)

	consoleReady := false
	for chunkPos, nbtList := range nbts {
		c := chunks[chunkPos]
		if c == nil {
			continue
		}
		for _, nbt := range nbtList {
			localPos, ok := nbtLocalPos(nbt)
			if !ok {
				continue
			}
			runtimeID := c.Block(uint8(floorMod(localPos.X(), 16)), int16(localPos.Y()), uint8(floorMod(localPos.Z(), 16)), 0)
			name, states, found := b.world.World().BlockRuntimeIDTable().RuntimeIDToState(runtimeID)
			if !found {
				continue
			}
			worldPos := b.localToWorldPos(localPos)
			if isCommandBlockName(name) {
				if b.IgnoreCommandBlock {
					continue
				}
				if err := b.buildCommandBlock(ctx, name, nbt, worldPos); err != nil {
					return fmt.Errorf("BuildTask.buildChunkGroupNBT: build command block %v: %w", worldPos, err)
				}
				b.publish(EventNameRunCommandBlockBuilt, worldPos)
				continue
			}

			if b.IgnoreOtherNBTBlock {
				continue
			}
			if !consoleReady {
				if err := b.initNBTConsole(ctx); err != nil {
					return fmt.Errorf("BuildTask.buildChunkGroupNBT: init nbt console: %w", err)
				}
				consoleReady = true
			}
			if err := b.placeNBTBlock(ctx, name, states, nbt, worldPos); err != nil {
				return fmt.Errorf("BuildTask.buildChunkGroupNBT: place nbt block %v: %w", worldPos, err)
			}
			b.publish(EventNameRunNBTBlockBuilt, worldPos)
		}
	}

	b.publish(EventNameRunNBTFinish, total)
	return nil
}

// countChunkGroupNBT 统计区块组中特殊方块写入阶段实际需要处理的数量。
func (b *BuildTask) countChunkGroupNBT(chunks map[define.ChunkPos]*chunk.Chunk, nbts map[define.ChunkPos][]map[string]any) int {
	total := 0
	for chunkPos, nbtList := range nbts {
		c := chunks[chunkPos]
		if c == nil {
			continue
		}
		for _, nbt := range nbtList {
			localPos, ok := nbtLocalPos(nbt)
			if !ok {
				continue
			}
			runtimeID := c.Block(uint8(floorMod(localPos.X(), 16)), int16(localPos.Y()), uint8(floorMod(localPos.Z(), 16)), 0)
			name, _, found := b.world.World().BlockRuntimeIDTable().RuntimeIDToState(runtimeID)
			if !found {
				continue
			}
			if isCommandBlockName(name) {
				if !b.IgnoreCommandBlock {
					total++
				}
				continue
			}
			if !b.IgnoreOtherNBTBlock {
				total++
			}
		}
	}
	return total
}

// buildCommandBlock 通过 CommandBlockUpdate 数据包写入命令方块配置。
func (b *BuildTask) buildCommandBlock(ctx context.Context, blockName string, nbt map[string]any, pos define.BlockPos) error {
	mode := uint32(packet_pb.CommandBlockEnum_CommandBlockImpulse)
	switch blockName {
	case "minecraft:chain_command_block":
		mode = uint32(packet_pb.CommandBlockEnum_CommandBlockChain)
	case "minecraft:repeating_command_block":
		mode = uint32(packet_pb.CommandBlockEnum_CommandBlockRepeating)
	}

	anyValue, err := anypb.New(&packet_pb.CommandBlockUpdate{
		Block:              true,
		Position:           blockPosToProto(pos),
		Mode:               mode,
		NeedsRedstone:      intValue(nbtValue(nbt, "auto", "Auto")) == 0,
		Conditional:        intValue(nbtValue(nbt, "conditionalMode", "ConditionalMode")) == 1,
		Command:            stringValue(nbt["Command"]),
		LastOutput:         stringValue(nbt["LastOutput"]),
		Name:               stringValue(nbt["CustomName"]),
		ShouldTrackOutput:  intValue(nbt["TrackOutput"]) == 1,
		TickDelay:          uint32(max(0, intValue(nbt["TickDelay"]))),
		ExecuteOnFirstTick: intValue(nbt["ExecuteOnFirstTick"]) == 1,
	})
	if err != nil {
		return fmt.Errorf("BuildTask.buildCommandBlock: pack packet: %w", err)
	}
	if err := b.frame.Client().Resources().WritePacket(ctx, &packet_pb.Packet{Value: anyValue}); err != nil {
		return fmt.Errorf("BuildTask.buildCommandBlock: write packet: %w", err)
	}
	return nil
}

// initNBTConsole 初始化 Core 侧 NBT 制作操作台。
func (b *BuildTask) initNBTConsole(ctx context.Context) error {
	pos := *b.ConsoleWorldPos
	return b.frame.Client().NBTAssigner().NewConsole(ctx, uint32(b.Dimension), blockPosToProto(pos))
}

// placeNBTBlock 使用 Core NBTAssigner 制作 NBT 方块，并搬运到目标位置。
func (b *BuildTask) placeNBTBlock(ctx context.Context, blockName string, states map[string]any, nbt map[string]any, pos define.BlockPos) error {
	result, err := b.frame.Client().NBTAssigner().PlaceNBTBlock(ctx, blockName, states, nbt)
	if err != nil {
		return fmt.Errorf("BuildTask.placeNBTBlock: make nbt block: %w", err)
	}

	if result.CanFast {
		stateStr := utils.PropertiesToStateStr(states)
		if err := b.sendSettingsCommand(ctx, fmt.Sprintf("setblock %d %d %d %s %s", pos.X(), pos.Y(), pos.Z(), blockName, stateStr), false); err != nil {
			return fmt.Errorf("BuildTask.placeNBTBlock: setblock: %w", err)
		}
		return nil
	}

	consolePos := *b.ConsoleWorldPos
	offset := result.Offset.GetValue()
	if len(offset) >= 3 {
		consolePos[0] += int(offset[0])
		consolePos[1] += int(offset[1])
		consolePos[2] += int(offset[2])
	}
	if err := b.saveTemporaryNBTStructure(ctx, consolePos); err != nil {
		return err
	}
	if err := b.sendWSCommandWithRetry(ctx, fmt.Sprintf(`structure load "%s" %d %d %d`, temporaryNBTStructureName, pos.X(), pos.Y(), pos.Z()), 60*time.Second, 1); err != nil {
		return fmt.Errorf("BuildTask.placeNBTBlock: structure load: %w", err)
	}
	_ = b.sendSettingsCommand(ctx, fmt.Sprintf(`structure delete "%s"`, temporaryNBTStructureName), false)
	return nil
}

// saveTemporaryNBTStructure 把 console 中制作好的 NBT 方块保存为内存结构。
func (b *BuildTask) saveTemporaryNBTStructure(ctx context.Context, pos define.BlockPos) error {
	command := fmt.Sprintf(
		`structure save "%s" %d %d %d %d %d %d false memory true`,
		temporaryNBTStructureName,
		pos.X(), pos.Y(), pos.Z(),
		pos.X(), pos.Y(), pos.Z(),
	)
	if err := b.sendWSCommandWithRetry(ctx, command, 60*time.Second, 1); err != nil {
		return fmt.Errorf("BuildTask.saveTemporaryNBTStructure: %w", err)
	}
	return nil
}

// sendWSCommandWithRetry 发送需要成功回执的 WS 命令，并按需重试。
func (b *BuildTask) sendWSCommandWithRetry(ctx context.Context, command string, timeout time.Duration, maxAttempts int) error {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, isTimeout, err := b.sendWSCommandWithTimeout(ctx, command, timeout)
		switch {
		case isTimeout:
			lastErr = fmt.Errorf("command timeout")
		case err != nil:
			lastErr = err
		case commandOutputSuccess(resp):
			return nil
		default:
			lastErr = fmt.Errorf("command failed")
		}
	}
	return lastErr
}

// commandOutputSuccess 判断命令回执是否表示成功。
func commandOutputSuccess(resp *packet_pb.CommandOutput) bool {
	if resp == nil {
		return false
	}
	if resp.GetSuccessCount() > 0 {
		return true
	}
	for _, message := range resp.GetOutputMessages() {
		if message.GetSuccess() {
			return true
		}
	}
	return false
}

// nbtLocalPos 读取 BuildingWorld 已经平移到构建区域局部坐标系的 NBT 坐标。
func nbtLocalPos(nbt map[string]any) (define.BlockPos, bool) {
	x, okX := intValueOK(nbt["x"])
	y, okY := intValueOK(nbt["y"])
	z, okZ := intValueOK(nbt["z"])
	if !okX || !okY || !okZ {
		return define.BlockPos{}, false
	}
	return define.BlockPos{x, y, z}, true
}

// localToWorldPos 将源结构局部坐标换算成目标世界坐标。
func (b *BuildTask) localToWorldPos(pos define.BlockPos) define.BlockPos {
	return define.BlockPos{
		b.StartPos.X() + pos.X(),
		b.StartPos.Y() + pos.Y(),
		b.StartPos.Z() + pos.Z(),
	}
}

// isCommandBlockName 判断方块名是否属于命令方块。
func isCommandBlockName(name string) bool {
	return strings.Contains(name, "command_block")
}

// blockPosToProto 将本地方块坐标转为协议坐标。
func blockPosToProto(pos define.BlockPos) *protocol.BlockPos {
	return &protocol.BlockPos{Value: []int32{int32(pos.X()), int32(pos.Y()), int32(pos.Z())}}
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}

func nbtValue(values map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := values[key]
		if ok {
			return value
		}
	}
	return nil
}

func intValue(value any) int {
	n, _ := intValueOK(value)
	return n
}

func intValueOK(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	case float32:
		return int(v), math.Trunc(float64(v)) == float64(v)
	case float64:
		return int(v), math.Trunc(v) == v
	default:
		return 0, false
	}
}

func floorMod(value, divisor int) int {
	result := value % divisor
	if result < 0 {
		result += divisor
	}
	return result
}
