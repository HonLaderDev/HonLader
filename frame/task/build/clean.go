package build

import (
	"context"
	"fmt"

	"github.com/HonLaderDev/HonLader/define"
	build_utils "github.com/HonLaderDev/HonLader/frame/task/build/utils"
)

// cleanChunkGroup 在构建当前区块组前清理目标区域中的已有方块。
func (b *BuildTask) cleanChunkGroup(ctx context.Context, groupPos define.ChunkPos) error {
	if !b.EnableAutoCleanBlock {
		return nil
	}

	start, end, ok := b.chunkGroupCleanBounds(groupPos)
	if !ok {
		return nil
	}
	commands := b.blockBuilder.BuildAirCommands(start, end)
	b.publish(EventNameRunCommandsGenerated, len(commands))
	for _, command := range commands {
		if err := b.sendSettingsCommand(ctx, command, false); err != nil {
			return fmt.Errorf("BuildTask.cleanChunkGroup: send clean command: %w", err)
		}
		b.publish(EventNameRunCommandSent, command)
	}
	return nil
}

func (b *BuildTask) chunkGroupCleanBounds(groupPos define.ChunkPos) (start, end define.BlockPos, ok bool) {
	size := b.world.Size()
	groupWidth := b.chunkGroupSide() * 16
	localStartX := int(groupPos.X()) * groupWidth
	localStartZ := int(groupPos.Z()) * groupWidth
	if localStartX >= size.Width || localStartZ >= size.Length {
		return define.BlockPos{}, define.BlockPos{}, false
	}

	localEndX := build_utils.MinInt(localStartX+groupWidth-1, size.Width-1)
	localEndZ := build_utils.MinInt(localStartZ+groupWidth-1, size.Length-1)
	start = define.BlockPos{
		b.StartPos.X() + localStartX,
		b.StartPos.Y(),
		b.StartPos.Z() + localStartZ,
	}
	end = define.BlockPos{
		b.StartPos.X() + localEndX,
		b.StartPos.Y() + size.Height - 1,
		b.StartPos.Z() + localEndZ,
	}
	return start, end, true
}
