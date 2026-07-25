package export

import "github.com/HonLaderDev/HonLader/frame"

const (
	// EventNameInitStart 初始化开始事件。
	// 参数：无。
	EventNameInitStart = Name + ".Init.Start"
	// EventNameInitOpenWorld 开始创建导出世界事件。
	// 参数：worldDir string。
	EventNameInitOpenWorld = Name + ".Init.OpenWorld"
	// EventNameInitFinish 初始化完成事件。
	// 参数：无。
	EventNameInitFinish = Name + ".Init.Finish"

	// EventNameRunStart 导出主流程开始事件。
	// 参数：size define.Size, total int。
	EventNameRunStart = Name + ".Run.Start"
	// EventNameRunBatchStart 批次开始处理事件。
	// 参数：index int。
	EventNameRunBatchStart = Name + ".Run.Batch.Start"
	// EventNameRunBatchMove 批次移动完成事件。
	// 参数：targetPos define.BlockPos。
	EventNameRunBatchMove = Name + ".Run.Batch.Move"
	// EventNameRunSubChunkRequest 子区块请求发送事件。
	// 参数：count int。
	EventNameRunSubChunkRequest = Name + ".Run.SubChunk.Request"
	// EventNameRunSubChunkTimeout 子区块请求超时事件。
	// 参数：count int。
	EventNameRunSubChunkTimeout = Name + ".Run.SubChunk.Timeout"
	// EventNameRunSubChunkDecodeFailed 子区块解码失败事件。
	// 参数：pos define.SubChunkPos, err error。
	EventNameRunSubChunkDecodeFailed = Name + ".Run.SubChunk.Decode.Failed"
	// EventNameRunSubChunkSaveFailed 子区块保存失败事件。
	// 参数：pos define.SubChunkPos, err error。
	EventNameRunSubChunkSaveFailed = Name + ".Run.SubChunk.Save.Failed"
	// EventNameRunChunkSaved 区块 NBT 保存完成事件。
	// 参数：pos define.ChunkPos。
	EventNameRunChunkSaved = Name + ".Run.Chunk.Saved"
	// EventNameRunChunkNBTSaveFailed 区块 NBT 保存失败事件。
	// 参数：pos define.ChunkPos, err error。
	EventNameRunChunkNBTSaveFailed = Name + ".Run.Chunk.NBT.Save.Failed"
	// EventNameRunBatchFinish 批次完成事件。
	// 参数：无。
	EventNameRunBatchFinish = Name + ".Run.Batch.Finish"
	// EventNameRunFinish 导出完成事件。
	// 参数：path string。
	EventNameRunFinish = Name + ".Run.Finish"
)

func (e *ExportTask) publish(name string, args ...any) {
	if name == EventNameRunBatchFinish {
		e.publishCheckpoint()
	}
	e.frame.EventBus().Publish(name, args...)
}

// publishCheckpoint 请求框架立即保存当前导出断点。
func (e *ExportTask) publishCheckpoint() {
	e.frame.EventBus().Publish(frame.EventNameTaskFrameTaskCheckpoint)
}
