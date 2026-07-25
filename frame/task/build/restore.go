package build

import (
	"fmt"

	"github.com/HonLaderDev/HonLader/define"
	task_codec "github.com/HonLaderDev/HonLader/frame/task"
)

// NewTaskFromInfo 通过持久化任务信息恢复构建任务。
func NewTaskFromInfo(info define.TaskInfo, frame define.TaskFrame) (define.Task, error) {
	task := new(BuildTask)
	if err := task_codec.UnmarshalTask(info, task); err != nil {
		return nil, fmt.Errorf("BuildTask.NewTaskFromInfo: %w", err)
	}
	task.frame = frame
	task.BuildTaskConfig.FillDefault()
	return task, nil
}
