package chunk_manager

// Progress 返回当前已处理的组数以及总组数。
func (c *ChunkManager) Progress() (int, int) {
	return c.progress, c.max
}

// Advance 将内部进度向前推进一组。
func (c *ChunkManager) Advance() {
	if c.progress < c.max {
		c.progress++
	}
}
