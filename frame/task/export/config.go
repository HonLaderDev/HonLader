package export

import "time"

const (
	defaultRequestTimeout = 3 * time.Second
	defaultRetryDelay     = 500 * time.Millisecond
	defaultChunkBatchSide = 16
)

// FillDefault 填充导出任务配置默认值。
func (c *ExportTaskConfig) FillDefault() {
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = defaultRequestTimeout
	}
	if c.RetryDelay <= 0 {
		c.RetryDelay = defaultRetryDelay
	}
	if c.ChunkBatchSide <= 0 {
		c.ChunkBatchSide = defaultChunkBatchSide
	}
}
