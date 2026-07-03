package frame

import (
	"fmt"
	"time"

	"github.com/RedLaderDev/RedLader/define"
	task_codec "github.com/RedLaderDev/RedLader/frame/task"
)

func (l *Launcher) watchTaskCheckpoints() {
	l.EventBus().SubscribeAsync(EventNameTaskFrameTaskStart, func(index int) {
		l.saveCheckpoint()
	}, false)
	l.EventBus().SubscribeAsync(EventNameTaskFrameTaskFinish, func(index int) {
		l.saveCheckpoint()
	}, false)
	l.EventBus().SubscribeAsync(EventNameTaskFrameTaskFailed, func(index int, err error) {
		l.saveCheckpoint()
	}, false)
	l.EventBus().SubscribeAsync(EventNameTaskFrameTaskCheckpoint, func() {
		l.saveCheckpoint()
	}, false)
	l.EventBus().SubscribeAsync(EventNameTaskFramePause, func(index int) {
		l.saveCheckpoint()
	}, false)
	l.EventBus().SubscribeAsync(EventNameTaskFrameStop, func(index int) {
		l.saveCheckpoint()
	}, false)
	l.EventBus().SubscribeAsync(EventNameTaskFrameClose, func() {
		l.saveCheckpoint()
	}, false)
}

func (l *Launcher) saveCheckpoint() {
	config, err := l.checkpointConfig()
	if err != nil {
		l.EventBus().Publish(EventNameLauncherCheckpointFailed, err)
		return
	}
	name := l.TaskGroupName()
	if name == "" {
		name = time.Now().Format("2006-01-02 15:04:05.000000000")
	}
	if err := l.dataManager.SaveCheckpointConfig(name, config); err != nil {
		l.EventBus().Publish(EventNameLauncherCheckpointFailed, err)
		return
	}
	l.EventBus().Publish(EventNameLauncherCheckpointSaved, config)
}

func (l *Launcher) checkpointConfig() (define.CheckpointConfig, error) {
	tasks := l.Tasks()
	taskInfos := make([]define.TaskInfo, 0, len(tasks))
	for i, task := range tasks {
		info, err := task_codec.MarshalTask(task)
		if err != nil {
			return define.CheckpointConfig{}, fmt.Errorf("Launcher.checkpointConfig: marshal task %d: %w", i, err)
		}
		taskInfos = append(taskInfos, info)
	}
	return define.CheckpointConfig{
		Tasks:            taskInfos,
		CurrentTaskIndex: l.CurrentTaskIndex(),
	}, nil
}
