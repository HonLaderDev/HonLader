package export

func (e *ExportTask) Stop() error {
	e.cancelTask()
	e.publishCheckpoint()
	return nil
}
