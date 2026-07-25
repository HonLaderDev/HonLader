package export

import (
	"context"
	"sync"
	"time"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/bedrock-world-operator/world"
)

// ExportTaskConfig 描述从当前服务器导出区域到 mcworld 的任务配置。
type ExportTaskConfig struct {
	// FilePath 导出的 mcworld 文件路径。
	FilePath string `config:"file_path"`
	// StartPos 导出区域起点。
	StartPos define.BlockPos `config:"start_pos"`
	// EndPos 导出区域终点。
	EndPos define.BlockPos `config:"end_pos"`
	// Dimension 导出区域所在维度。
	Dimension define.Dimension `config:"dimension"`
	// RequestTimeout 单次子区块请求超时时间。
	RequestTimeout time.Duration `config:"request_timeout"`
	// RetryDelay 子区块请求超时后的重试等待时间。
	RetryDelay time.Duration `config:"retry_delay"`
	// ChunkBatchSide 每次请求的区块批次边长。
	ChunkBatchSide int `config:"chunk_batch_side"`
}

// ExportTaskCheckpoint 记录导出任务断点。
type ExportTaskCheckpoint struct {
	// CurrentBatch 当前已完成的导出批次序号。
	CurrentBatch int `checkpoint:"current_batch"`
}

// ExportTask 从服务器导出指定区域并保存为 mcworld。
type ExportTask struct {
	ExportTaskConfig     `task:"config"`
	ExportTaskCheckpoint `task:"checkpoint"`

	frame      define.TaskFrame
	world      *world.BedrockWorld
	tempDir    string
	worldDir   string
	outputPath string

	taskMu     sync.Mutex
	taskCtx    context.Context
	taskCancel context.CancelFunc
}

func (c ExportTaskConfig) NewTask(frame define.TaskFrame) define.Task {
	c.FillDefault()
	task := new(ExportTask)
	task.ExportTaskConfig = c
	task.frame = frame
	return task
}
