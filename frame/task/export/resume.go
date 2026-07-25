package export

import "fmt"

// Resume 初始化任务并从断点继续执行导出任务。
func (e *ExportTask) Resume() error {
	if err := e.Init(); err != nil {
		return fmt.Errorf("ExportTask.Resume: %w", err)
	}
	ctx := e.startTaskContext()
	if err := e.run(ctx); err != nil {
		return fmt.Errorf("ExportTask.Resume: %w", err)
	}
	return nil
}
