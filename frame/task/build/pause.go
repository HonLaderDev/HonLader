package build

func (b *BuildTask) Pause() error {
	b.cancelTask()
	b.publishCheckpoint()
	return nil
}
