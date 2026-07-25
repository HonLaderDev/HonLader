package export

import (
	"fmt"

	"github.com/HonLaderDev/HonLader/define"
	task_codec "github.com/HonLaderDev/HonLader/frame/task"
)

// NewTaskFromInfo 通过持久化任务信息恢复导出任务。
func NewTaskFromInfo(info define.TaskInfo, frame define.TaskFrame) (define.Task, error) {
	task := new(ExportTask)
	if err := task_codec.UnmarshalTask(info, task); err != nil {
		return nil, fmt.Errorf("ExportTask.NewTaskFromInfo: %w", err)
	}
	task.frame = frame
	task.ExportTaskConfig.FillDefault()
	return task, nil
}
