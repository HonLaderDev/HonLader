package build

import "github.com/RedLaderDev/RedLader/frame"

const (
	// EventNameInitStart 初始化开始事件。
	// 参数：无。
	EventNameInitStart = Name + ".Init.Start"
	// EventNameInitOpenWorld 开始打开源世界事件。
	// 参数：worldPath string。
	EventNameInitOpenWorld = Name + ".Init.OpenWorld"
	// EventNameInitFinish 初始化完成事件。
	// 参数：无。
	EventNameInitFinish = Name + ".Init.Finish"

	// EventNameRunStart 构建主流程开始事件。
	// 参数：size define.Size, total int。
	EventNameRunStart = Name + ".Run.Start"
	// EventNameRunTickingAreaCheckStart 常加载区域容量检查开始事件。
	// 参数：无。
	EventNameRunTickingAreaCheckStart = Name + ".Run.TickingArea.Check.Start"
	// EventNameRunTickingAreaCheckFinish 常加载区域容量检查完成事件。
	// 参数：available int, required int。
	EventNameRunTickingAreaCheckFinish = Name + ".Run.TickingArea.Check.Finish"
	// EventNameRunTickingAreaDisabled 常加载区域运行期降级关闭事件。
	// 参数：available int, required int, reason string。
	EventNameRunTickingAreaDisabled = Name + ".Run.TickingArea.Disabled"
	// EventNameRunChunkGroupStart 区块组开始处理事件。
	// 参数：progress int。
	EventNameRunChunkGroupStart = Name + ".Run.ChunkGroup.Start"
	// EventNameRunChunkGroupMove 区块组机器人移动完成事件。
	// 参数：groupPos define.ChunkPos, targetPos define.BlockPos。
	EventNameRunChunkGroupMove = Name + ".Run.ChunkGroup.Move"
	// EventNameRunChunkGroupPreHandleStart 下一个区块组预读取开始事件。
	// 参数：index int, groupPos define.ChunkPos。
	EventNameRunChunkGroupPreHandleStart = Name + ".Run.ChunkGroup.PreHandle.Start"
	// EventNameRunChunkGroupPreHandleFinish 下一个区块组预读取完成事件。
	// 参数：index int, groupPos define.ChunkPos, chunkCount int, nbtCount int, err error。
	EventNameRunChunkGroupPreHandleFinish = Name + ".Run.ChunkGroup.PreHandle.Finish"
	// EventNameRunChunkGroupPreWaitStart 下一个区块组预等待加载开始事件。
	// 参数：index int, groupPos define.ChunkPos。
	EventNameRunChunkGroupPreWaitStart = Name + ".Run.ChunkGroup.PreWait.Start"
	// EventNameRunChunkGroupPreWaitFinish 下一个区块组预等待加载完成事件。
	// 参数：index int, groupPos define.ChunkPos, err error。
	EventNameRunChunkGroupPreWaitFinish = Name + ".Run.ChunkGroup.PreWait.Finish"
	// EventNameRunTickingAreaWaitStart 等待常加载区域槽位开始事件。
	// 参数：name string。
	EventNameRunTickingAreaWaitStart = Name + ".Run.TickingArea.Wait.Start"
	// EventNameRunTickingAreaWaitFinish 等待常加载区域槽位完成事件。
	// 参数：name string。
	EventNameRunTickingAreaWaitFinish = Name + ".Run.TickingArea.Wait.Finish"
	// EventNameRunTickingAreaAddStart 常加载区域创建开始事件。
	// 参数：groupPos define.ChunkPos, name string。
	EventNameRunTickingAreaAddStart = Name + ".Run.TickingArea.Add.Start"
	// EventNameRunTickingAreaAddFinish 常加载区域创建命令发送完成事件。
	// 参数：groupPos define.ChunkPos, name string。
	EventNameRunTickingAreaAddFinish = Name + ".Run.TickingArea.Add.Finish"
	// EventNameRunTickingAreaRemoveStart 常加载区域删除开始事件。
	// 参数：name string。
	EventNameRunTickingAreaRemoveStart = Name + ".Run.TickingArea.Remove.Start"
	// EventNameRunTickingAreaRemoveFinish 常加载区域删除完成事件。
	// 参数：name string。
	EventNameRunTickingAreaRemoveFinish = Name + ".Run.TickingArea.Remove.Finish"
	// EventNameRunTickingAreaRemoveFailed 常加载区域删除失败事件。
	// 参数：name string, err error。
	EventNameRunTickingAreaRemoveFailed = Name + ".Run.TickingArea.Remove.Failed"
	// EventNameRunChunkGroupWaitLoadStart 区块组等待加载开始事件。
	// 参数：groupPos define.ChunkPos。
	EventNameRunChunkGroupWaitLoadStart = Name + ".Run.ChunkGroup.WaitLoad.Start"
	// EventNameRunChunkGroupWaitLoadProbe 区块组加载探测事件。
	// 参数：groupPos define.ChunkPos, attempt int, ready bool, timeout bool, message string。
	EventNameRunChunkGroupWaitLoadProbe = Name + ".Run.ChunkGroup.WaitLoad.Probe"
	// EventNameRunChunkGroupWaitLoadRetry 区块组加载探测失败后准备重试事件。
	// 参数：groupPos define.ChunkPos, attempt int, timeout bool, message string。
	EventNameRunChunkGroupWaitLoadRetry = Name + ".Run.ChunkGroup.WaitLoad.Retry"
	// EventNameRunChunkGroupWaitLoadFinish 区块组等待加载完成事件。
	// 参数：groupPos define.ChunkPos。
	EventNameRunChunkGroupWaitLoadFinish = Name + ".Run.ChunkGroup.WaitLoad.Finish"
	// EventNameRunChunkGroupLoaded 区块组数据读取完成事件。
	// 参数：chunks map[define.ChunkPos]*chunk.Chunk, nbts map[define.ChunkPos][]map[string]any。
	EventNameRunChunkGroupLoaded = Name + ".Run.ChunkGroup.Loaded"
	// EventNameRunCommandsGenerated 区块组构建命令生成完成事件。
	// 参数：commandCount int。
	EventNameRunCommandsGenerated = Name + ".Run.Commands.Generated"
	// EventNameRunCommandSent 单条构建命令发送完成事件。
	// 参数：command string。
	EventNameRunCommandSent = Name + ".Run.Command.Sent"
	// EventNameRunItemCleanStart 区块组掉落物清理开始事件。
	// 参数：groupPos define.ChunkPos, command string。
	EventNameRunItemCleanStart = Name + ".Run.ItemClean.Start"
	// EventNameRunItemCleanFinish 区块组掉落物清理完成事件。
	// 参数：groupPos define.ChunkPos, command string。
	EventNameRunItemCleanFinish = Name + ".Run.ItemClean.Finish"
	// EventNameRunChunkGroupFinish 区块组处理完成事件。
	// 参数：无。
	EventNameRunChunkGroupFinish = Name + ".Run.ChunkGroup.Finish"
	// EventNameRunFinish 构建主流程完成事件。
	// 参数：无。
	EventNameRunFinish = Name + ".Run.Finish"
)

// publish 向任务所属框架发布构建事件。
func (b *BuildTask) publish(name string, args ...any) {
	if name == EventNameRunChunkGroupFinish {
		b.frame.EventBus().Publish(frame.EventNameTaskFrameTaskCheckpoint)
	}
	b.frame.EventBus().Publish(name, args...)
}
