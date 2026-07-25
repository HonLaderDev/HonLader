package build

import (
	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils"
)

// FillDefault 填充构建任务配置的默认值。
func (c *BuildTaskConfig) FillDefault() {
	if c.Speed == nil {
		c.Speed = utils.Ptr(defaultSpeed)
	}
	if c.ChunkGroupSide == nil {
		c.ChunkGroupSide = utils.Ptr(defaultChunkGroupSide)
	}
	if c.FixModeTimeout == nil {
		c.FixModeTimeout = utils.Ptr(10.0)
	}
	if c.ConsoleWorldPos == nil {
		c.ConsoleWorldPos = utils.Ptr(define.BlockPos{-50, 0, -50})
	}
	if c.GameProgressRefreshDelay == nil {
		c.GameProgressRefreshDelay = utils.Ptr(0.5)
	}
}
