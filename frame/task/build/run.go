package build

import (
	"context"
	"fmt"
)

// run 执行构建任务主流程。
func (b *BuildTask) run(ctx context.Context) error {
	// 退出时清理任务上下文。
	defer b.finishTaskContext(ctx)

	// 退出时清理常加载区域。
	defer func() {
		_ = b.releaseAllTickingAreas(context.Background())
	}()

	// 初始化常加载状态。
	if err := b.prepareTickingAreaRuntime(ctx); err != nil {
		return fmt.Errorf("BuildTask.run: prepare tickingarea state: %w", err)
	}

	// 读取初始构建进度。
	progress, total := b.chunkManager.Progress()
	b.publish(EventNameRunStart, b.world.Size(), total)

	var currentFuture *chunkGroupFuture
	for ; progress < total; progress++ {
		// 标记当前区块组。
		groupPos := b.chunkManager.ChunkGroupPos(progress)
		b.publish(EventNameRunChunkGroupStart, progress)

		// 预读取下一组数据。
		nextIndex := progress + 1
		var nextFuture *chunkGroupFuture
		if nextIndex < total {
			if b.preHandleNextChunkGroup {
				nextFuture = b.startPreHandleNextChunkGroup(ctx, nextIndex)
			}
		}

		// 移动到区块组中心。
		targetPos, err := b.moveBotToChunkGroup(ctx, groupPos)
		if err != nil {
			return fmt.Errorf("BuildTask.run: move bot to chunk group: %w", err)
		}
		b.publish(EventNameRunChunkGroupMove, groupPos, targetPos)

		// 确认区块组已加载。
		if err := b.ensureChunkGroupLoad(ctx, progress, groupPos); err != nil {
			return fmt.Errorf("BuildTask.run: ensure chunk group load: %w", err)
		}

		// 预等待下一组加载。
		b.preloadNextChunkGroupLoad(ctx, nextIndex, total)

		// 清理目标区域方块。
		if err := b.cleanChunkGroup(ctx, groupPos); err != nil {
			return fmt.Errorf("BuildTask.run: clean chunk group: %w", err)
		}

		// 读取当前区块组数据。
		data, err := b.loadCurrentChunkGroup(ctx, progress, currentFuture)
		if err != nil {
			return fmt.Errorf("BuildTask.run: load current chunk group: %w", err)
		}
		b.publish(EventNameRunChunkGroupLoaded, data.chunks, data.nbts)

		// 生成普通方块命令。
		commands := b.blockBuilder.BuildCommands(data.chunks)
		b.publish(EventNameRunCommandsGenerated, len(commands))

		// 发送构建命令。
		for _, command := range commands {
			if err := b.sendSettingsCommand(ctx, command, false); err != nil {
				return fmt.Errorf("BuildTask.run: send build command: %w", err)
			}
			b.publish(EventNameRunCommandSent, command)
		}

		// 清理区块组掉落物。
		if err := b.cleanChunkGroupItems(ctx, groupPos); err != nil {
			return fmt.Errorf("BuildTask.run: clean chunk group items: %w", err)
		}

		// 推进持久化断点。
		b.updateCurrentChunk(nextIndex)
		b.publish(EventNameRunChunkGroupFinish)

		// 释放上一组常加载。
		if err := b.releasePreviousTickingArea(ctx); err != nil {
			return fmt.Errorf("BuildTask.run: release previous tickingarea: %w", err)
		}

		// 切换下一组预读结果。
		currentFuture = nextFuture
	}

	// 发布构建完成事件。
	b.publish(EventNameRunFinish)
	return nil
}

// updateCurrentChunk 将当前区块组进度换算回区块断点，便于任务恢复。
func (b *BuildTask) updateCurrentChunk(progress int) {
	b.CurrentChunk = progress * b.chunkGroupSide() * b.chunkGroupSide()
}
