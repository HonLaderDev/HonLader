package export

import "fmt"

// Start 初始化任务并从当前断点开始执行导出任务。
func (e *ExportTask) Start() error {
	if err := e.Init(); err != nil {
		return fmt.Errorf("ExportTask.Start: %w", err)
	}
	ctx := e.startTaskContext()
	if err := e.run(ctx); err != nil {
		return fmt.Errorf("ExportTask.Start: %w", err)
	}
	return nil
}
