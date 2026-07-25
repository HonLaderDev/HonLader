package build

import (
	"context"
	"errors"
	"fmt"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/bedrock-world-operator/chunk"
)

// chunkGroupData 保存一个区块组读取完成后的构建输入。
type chunkGroupData struct {
	// index 是区块组在 ChunkManager 遍历序列中的序号。
	index int
	// groupPos 是区块组坐标，不是普通区块坐标。
	groupPos define.ChunkPos
	// chunks 保存当前区块组内实际存在的区块数据。
	chunks map[define.ChunkPos]*chunk.Chunk
	// nbts 保存当前区块组内读取到的方块实体 NBT 数据。
	nbts map[define.ChunkPos][]map[string]any
}

func (d chunkGroupData) empty() bool {
	return len(d.chunks) == 0 && len(d.nbts) == 0
}

// chunkGroupFuture 表示一个后台区块组读取任务。
type chunkGroupFuture struct {
	// data 是后台读取成功后的区块组数据。
	data chunkGroupData
	// err 是后台读取失败时保留的错误。
	err error
	// done 在后台读取完成后关闭，用于通知等待方。
	done chan struct{}
}

// chunkGroupPreWaitFuture 表示一个后台区块加载预等待任务。
type chunkGroupPreWaitFuture struct {
	// err 是后台预等待失败时保留的错误。
	err error
	// done 在后台预等待完成后关闭，用于通知等待方。
	done chan struct{}
}

func newChunkGroupFuture() *chunkGroupFuture {
	return &chunkGroupFuture{done: make(chan struct{})}
}

func (f *chunkGroupFuture) resolve(data chunkGroupData, err error) {
	f.data = data
	f.err = err
	close(f.done)
}

func (f *chunkGroupFuture) wait(ctx context.Context) (chunkGroupData, error) {
	select {
	case <-f.done:
		if f.err != nil {
			return chunkGroupData{}, fmt.Errorf("chunkGroupFuture.wait: %w", f.err)
		}
		return f.data, nil
	case <-ctx.Done():
		return chunkGroupData{}, fmt.Errorf("chunkGroupFuture.wait: %w", ctx.Err())
	}
}

func newChunkGroupPreWaitFuture() *chunkGroupPreWaitFuture {
	return &chunkGroupPreWaitFuture{done: make(chan struct{})}
}

func (f *chunkGroupPreWaitFuture) resolve(err error) {
	f.err = err
	close(f.done)
}

func (f *chunkGroupPreWaitFuture) wait(ctx context.Context) error {
	select {
	case <-f.done:
		if f.err != nil {
			return fmt.Errorf("chunkGroupPreWaitFuture.wait: %w", f.err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("chunkGroupPreWaitFuture.wait: %w", ctx.Err())
	}
}

// startPreHandleNextChunkGroup 在后台提前读取下一组区块和 NBT 数据。
func (b *BuildTask) startPreHandleNextChunkGroup(ctx context.Context, index int) *chunkGroupFuture {
	future := newChunkGroupFuture()
	groupPos := b.chunkManager.ChunkGroupPos(index)
	b.publish(EventNameRunChunkGroupPreHandleStart, index, groupPos)
	go func() {
		if err := ctx.Err(); err != nil {
			future.resolve(chunkGroupData{}, fmt.Errorf("BuildTask.startPreHandleNextChunkGroup: %w", err))
			b.publish(EventNameRunChunkGroupPreHandleFinish, index, groupPos, 0, 0, err)
			return
		}
		chunks, nbts, err := b.chunkManager.ChunkGroup(index)
		if err != nil {
			err = fmt.Errorf("BuildTask.startPreHandleNextChunkGroup: %w", err)
		}
		data := chunkGroupData{
			index:    index,
			groupPos: groupPos,
			chunks:   chunks,
			nbts:     nbts,
		}
		future.resolve(data, err)
		b.publish(EventNameRunChunkGroupPreHandleFinish, index, groupPos, len(chunks), len(nbts), err)
	}()
	return future
}

// startPreWaitNextChunkLoad 在后台提前等待下一组区块可访问。
func (b *BuildTask) startPreWaitNextChunkLoad(ctx context.Context, index int, groupPos define.ChunkPos) *chunkGroupPreWaitFuture {
	future := newChunkGroupPreWaitFuture()
	var noErr error
	b.publish(EventNameRunChunkGroupPreWaitStart, index, groupPos)
	go func() {
		if b.preWaitNextChunkTickingArea {
			if err := b.prepareChunkLoadTickingArea(ctx, groupPos); err != nil {
				b.publish(EventNameRunChunkGroupPreWaitFinish, index, groupPos, err)
				future.resolve(fmt.Errorf("BuildTask.startPreWaitNextChunkLoad: prepare tickingarea: %w", err))
				return
			}
		}
		if err := b.waitChunkLoad(ctx, groupPos); err != nil {
			b.publish(EventNameRunChunkGroupPreWaitFinish, index, groupPos, err)
			future.resolve(fmt.Errorf("BuildTask.startPreWaitNextChunkLoad: wait chunk load: %w", err))
			return
		}
		b.publish(EventNameRunChunkGroupPreWaitFinish, index, groupPos, noErr)
		future.resolve(nil)
	}()
	return future
}

func (b *BuildTask) waitPreWaitNextChunkLoad(ctx context.Context, future *chunkGroupPreWaitFuture) error {
	if future == nil {
		return nil
	}
	if err := future.wait(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("BuildTask.waitPreWaitNextChunkLoad: %w", err)
		}
		return err
	}
	return nil
}

// ensureChunkGroupLoad 确认当前区块组已经可被服务端访问。
//
// 如果上一轮已经给当前组启动了预等待，则优先消费预等待结果；预等待失败但任务上下文未取消时，
// 会退回到同步常加载区域准备和 fill keep 探测，避免因为一次后台探测失败跳过加载检查。
func (b *BuildTask) ensureChunkGroupLoad(ctx context.Context, index int, groupPos define.ChunkPos) error {
	if b.preWaitChunkGroupFuture != nil && b.preWaitChunkGroupIndex == index {
		future := b.preWaitChunkGroupFuture
		b.preWaitChunkGroupIndex = -1
		b.preWaitChunkGroupFuture = nil
		if err := b.waitPreWaitNextChunkLoad(ctx, future); err == nil {
			return nil
		} else if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("BuildTask.ensureChunkGroupLoad: wait pre wait chunk load: %w", err)
		}
	}

	if err := b.prepareChunkLoadTickingArea(ctx, groupPos); err != nil {
		return fmt.Errorf("BuildTask.ensureChunkGroupLoad: prepare chunk load tickingarea: %w", err)
	}
	if err := b.waitChunkLoad(ctx, groupPos); err != nil {
		return fmt.Errorf("BuildTask.ensureChunkGroupLoad: wait chunk load: %w", err)
	}
	return nil
}

// preloadNextChunkGroupLoad 给下一组启动后台加载探测。
func (b *BuildTask) preloadNextChunkGroupLoad(ctx context.Context, index int, total int) {
	if index >= total || !b.preWaitNextChunkLoad {
		return
	}
	b.preWaitChunkGroupIndex = index
	b.preWaitChunkGroupFuture = b.startPreWaitNextChunkLoad(ctx, index, b.chunkManager.ChunkGroupPos(index))
}

// loadCurrentChunkGroup 获取当前组数据；如果预读取命中，则直接消费预读取结果。
func (b *BuildTask) loadCurrentChunkGroup(ctx context.Context, progress int, future *chunkGroupFuture) (chunkGroupData, error) {
	if future != nil {
		data, err := future.wait(ctx)
		if err != nil {
			return chunkGroupData{}, fmt.Errorf("BuildTask.loadCurrentChunkGroup: wait pre handle: %w", err)
		}
		if data.index == progress {
			b.chunkManager.Advance()
			return data, nil
		}
	}

	chunks, nbts, err := b.chunkManager.NextChunkGroup()
	if err != nil {
		return chunkGroupData{}, fmt.Errorf("BuildTask.loadCurrentChunkGroup: next chunk group: %w", err)
	}
	return chunkGroupData{
		index:    progress,
		groupPos: b.chunkManager.ChunkGroupPos(progress),
		chunks:   chunks,
		nbts:     nbts,
	}, nil
}
