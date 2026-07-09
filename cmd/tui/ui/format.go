package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/HonLaderDev/HonLader/define"
)

type TextUIFormat struct{}

func (TextUIFormat) String(value string) string {
	if strings.TrimSpace(value) == "" {
		return "未设置"
	}
	return value
}

func (TextUIFormat) BlockPos(value define.BlockPos) string {
	return fmt.Sprintf("%d,%d,%d", value.X(), value.Y(), value.Z())
}

func (TextUIFormat) Dimension(value define.Dimension) string {
	switch value {
	case define.DimensionIDOverworld:
		return "主世界"
	case define.DimensionIDNether:
		return "下界"
	case define.DimensionIDEnd:
		return "末地"
	}
	return fmt.Sprintf("%d", value)
}

func (TextUIFormat) Bool(value bool) string {
	if value {
		return "开"
	}
	return "关"
}

func (TextUIFormat) Float(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (f TextUIFormat) FloatSecond(value float64) string {
	return f.Float(value) + "s"
}

func (f TextUIFormat) OptionalString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
