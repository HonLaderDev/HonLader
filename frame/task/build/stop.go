package build

func (b *BuildTask) Stop() error {
	b.cancelTask()
	b.publishCheckpoint()
	return nil
}
