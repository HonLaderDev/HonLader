package build

import (
	"context"
	"fmt"

	"github.com/RedLaderDev/Fatalder/define"
)

// cleanChunkGroupItems 清理当前区块组目标范围内的掉落物。
func (b *BuildTask) cleanChunkGroupItems(ctx context.Context, groupPos define.ChunkPos) error {
	if b.DisableAutoCleanItem {
		return nil
	}

	startX, _, startZ, endX, endZ := b.chunkLoadBounds(groupPos)
	command := fmt.Sprintf(
		"kill @e[type=Item,x=%d,y=%d,z=%d,dx=%d,dy=%d,dz=%d]",
		startX,
		define.WorldRange[0],
		startZ,
		endX-startX+1,
		define.WorldRange[1]-define.WorldRange[0]+1,
		endZ-startZ+1,
	)
	b.publish(EventNameRunItemCleanStart, groupPos, command)
	if err := b.sendSettingsCommand(ctx, command, false); err != nil {
		return fmt.Errorf("BuildTask.cleanChunkGroupItems: %w", err)
	}
	b.publish(EventNameRunItemCleanFinish, groupPos, command)
	return nil
}
