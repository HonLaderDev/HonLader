package export

func (e *ExportTask) Pause() error {
	e.cancelTask()
	e.publishCheckpoint()
	return nil
}
