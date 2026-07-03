package frame

const (
	// EventNameTaskFrameTaskAdded 添加任务事件。
	// 参数：index int。
	EventNameTaskFrameTaskAdded = "TaskFrame.Task.Added"

	// EventNameTaskFrameStart 任务框架启动事件。
	// 参数：total int。
	EventNameTaskFrameStart = "TaskFrame.Start"
	// EventNameTaskFrameFinish 任务框架完成事件。
	// 参数：无。
	EventNameTaskFrameFinish = "TaskFrame.Finish"

	// EventNameTaskFrameTaskStart 单个任务启动事件。
	// 参数：index int。
	EventNameTaskFrameTaskStart = "TaskFrame.Task.Start"
	// EventNameTaskFrameTaskFinish 单个任务完成事件。
	// 参数：index int。
	EventNameTaskFrameTaskFinish = "TaskFrame.Task.Finish"
	// EventNameTaskFrameTaskFailed 单个任务失败事件。
	// 参数：index int, err error。
	EventNameTaskFrameTaskFailed = "TaskFrame.Task.Failed"
	// EventNameTaskFrameTaskCheckpoint 请求保存单个任务断点事件。
	// 参数：无。
	EventNameTaskFrameTaskCheckpoint = "TaskFrame.Task.Checkpoint"
	// EventNameTaskFrameNextTask 准备运行下一个任务事件。
	// 参数：index int。
	EventNameTaskFrameNextTask = "TaskFrame.Task.Next"

	// EventNameTaskFramePause 暂停当前任务事件。
	// 参数：index int。
	EventNameTaskFramePause = "TaskFrame.Pause"
	// EventNameTaskFrameResume 恢复当前任务事件。
	// 参数：index int。
	EventNameTaskFrameResume = "TaskFrame.Resume"
	// EventNameTaskFrameStop 停止当前任务事件。
	// 参数：index int。
	EventNameTaskFrameStop = "TaskFrame.Stop"
	// EventNameTaskFrameClose 关闭任务框架事件。
	// 参数：无。
	EventNameTaskFrameClose = "TaskFrame.Close"

	// EventNameLauncherCheckpointSaved 自动保存断点完成事件。
	// 参数：config define.CheckpointConfig。
	EventNameLauncherCheckpointSaved = "Launcher.Checkpoint.Saved"
	// EventNameLauncherCheckpointFailed 自动保存断点失败事件。
	// 参数：err error。
	EventNameLauncherCheckpointFailed = "Launcher.Checkpoint.Failed"
)

func (f *TaskFrame) publish(name string, args ...any) {
	f.eventBus.Publish(name, args...)
}
