package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/HonLaderDev/HonLader/define"
)

// Format 统一格式化 TUI 摘要和菜单中的展示值。
func (t *TextUI) Format(value any) string {
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "未设置"
		}
		return v
	case bool:
		if v {
			return "开"
		}
		return "关"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case define.BlockPos:
		return fmt.Sprintf("%d,%d,%d", v.X(), v.Y(), v.Z())
	case define.Dimension:
		switch v {
		case define.DimensionIDOverworld:
			return "主世界"
		case define.DimensionIDNether:
			return "下界"
		case define.DimensionIDEnd:
			return "末地"
		default:
			return fmt.Sprintf("%d", v)
		}
	default:
		return fmt.Sprint(value)
	}
}
