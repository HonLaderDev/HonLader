package define

// Size 表示三维区域尺寸。
type Size struct {
	// Width 是 X 轴方向长度。
	Width int
	// Height 是 Y 轴方向高度。
	Height int
	// Length 是 Z 轴方向长度。
	Length int
}

// Volume 返回三维区域的体积。
func (s *Size) Volume() int {
	return s.Width * s.Height * s.Length
}

// ChunkXCount 返回 X 轴方向覆盖的区块数量。
func (s *Size) ChunkXCount() int {
	return (s.Width + 16 - 1) / 16
}

// ChunkZCount 返回 Z 轴方向覆盖的区块数量。
func (s *Size) ChunkZCount() int {
	return (s.Length + 16 - 1) / 16
}

// ChunkCount 返回 X/Z 平面覆盖的区块总数。
func (s *Size) ChunkCount() int {
	return s.ChunkXCount() * s.ChunkZCount()
}
