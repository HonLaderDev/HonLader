package build_task_tui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/HonLaderDev/HonLader/define"
)

// 建筑文件、建筑区域、源维度

// PromptBuildTaskWorldPath 读取建筑文件路径。
func (t *BuildTaskTUI) PromptBuildTaskWorldPath(ctx context.Context, currentValues ...string) (string, error) {
	current := ""
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	files := CurrentDirectoryFiles()
	options := []string{"自定义文件路径"}
	for _, file := range files {
		options = append(options, fmt.Sprintf("%s(%s)", file.Name, FormatFileSize(file.Size)))
	}
	sel, err := t.tui.Control().Select(ctx, "请选择建筑文件：\n", options)
	if err != nil {
		return "", err
	}
	if sel > 0 {
		return files[sel-1].Name, nil
	}
	if len(currentValues) > 0 && currentValues[0] != "" {
		return t.tui.PromptRequired(ctx, fmt.Sprintf("请输入建筑文件路径(当前 %s)：", current), current)
	}
	return t.tui.PromptRequired(ctx, "请输入建筑文件路径：")
}

// DirectoryFile 描述当前目录普通文件。
type DirectoryFile struct {
	Name string
	Size int64
}

// CurrentDirectoryFiles 返回当前目录下的普通文件。
func CurrentDirectoryFiles() []DirectoryFile {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil
	}
	files := make([]DirectoryFile, 0, len(entries))
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, DirectoryFile{Name: entry.Name(), Size: info.Size()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return files
}

// FormatFileSize 格式化文件大小。
func FormatFileSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d%s", size, units[unit])
	}
	return fmt.Sprintf("%.2f%s", value, units[unit])
}

// PromptBuildTaskWorldStartPos 读取建筑区域起点坐标。
func (t *BuildTaskTUI) PromptBuildTaskWorldStartPos(ctx context.Context, currentValues ...define.BlockPos) (define.BlockPos, error) {
	if len(currentValues) > 0 {
		return t.tui.PromptBlockPos(ctx, fmt.Sprintf("请输入建筑区域起点(当前 %s)：", t.tui.Format(currentValues[0])), currentValues[0])
	}
	return t.tui.PromptBlockPos(ctx, "请输入建筑区域起点(x,y,z)：")
}

// PromptBuildTaskWorldEndPos 读取建筑区域终点坐标。
func (t *BuildTaskTUI) PromptBuildTaskWorldEndPos(ctx context.Context, currentValues ...define.BlockPos) (define.BlockPos, error) {
	if len(currentValues) > 0 {
		return t.tui.PromptBlockPos(ctx, fmt.Sprintf("请输入建筑区域终点(当前 %s)：", t.tui.Format(currentValues[0])), currentValues[0])
	}
	return t.tui.PromptBlockPos(ctx, "请输入建筑区域终点(x,y,z)：")
}

// PromptBuildTaskWorldDimension 读取源建筑区域所在维度。
func (t *BuildTaskTUI) PromptBuildTaskWorldDimension(ctx context.Context, args ...any) (define.Dimension, error) {
	title, current, hasCurrent := BuildTaskDimensionPromptArgs("请选择建筑区域维度：\n", args...)
	return t.PromptBuildTaskDimensionMenu(ctx, title, current, hasCurrent)
}

// 构建起点、目标维度

// PromptBuildTaskStartPos 读取目标世界中的构建起点坐标。
func (t *BuildTaskTUI) PromptBuildTaskStartPos(ctx context.Context, currentValues ...define.BlockPos) (define.BlockPos, error) {
	if len(currentValues) > 0 {
		return t.tui.PromptBlockPos(ctx, fmt.Sprintf("请输入构建起点(当前 %s)：", t.tui.Format(currentValues[0])), currentValues[0])
	}
	return t.tui.PromptBlockPos(ctx, "请输入构建起点(x,y,z)：")
}

// PromptBuildTaskDimension 读取目标构建维度。
func (t *BuildTaskTUI) PromptBuildTaskDimension(ctx context.Context, args ...any) (define.Dimension, error) {
	title, current, hasCurrent := BuildTaskDimensionPromptArgs("请选择构建维度：\n", args...)
	return t.PromptBuildTaskDimensionMenu(ctx, title, current, hasCurrent)
}

// BuildTaskDimensionPromptArgs 解析维度选择菜单的标题和当前值参数。
func BuildTaskDimensionPromptArgs(defaultTitle string, args ...any) (string, define.Dimension, bool) {
	title := defaultTitle
	var current define.Dimension
	hasCurrent := false
	for _, arg := range args {
		switch value := arg.(type) {
		case string:
			title = value
		case define.Dimension:
			current = value
			hasCurrent = true
		}
	}
	return title, current, hasCurrent
}

// PromptBuildTaskDimensionMenu 展示维度选择菜单并返回选择结果。
func (t *BuildTaskTUI) PromptBuildTaskDimensionMenu(ctx context.Context, title string, current define.Dimension, hasCurrent bool) (define.Dimension, error) {
	options := []string{"主世界", "下界", "末地", "自定义"}
	if hasCurrent {
		options = []string{
			fmt.Sprintf("%s (%s)", "主世界", CurrentDimensionMarker(current, define.DimensionIDOverworld)),
			fmt.Sprintf("%s (%s)", "下界", CurrentDimensionMarker(current, define.DimensionIDNether)),
			fmt.Sprintf("%s (%s)", "末地", CurrentDimensionMarker(current, define.DimensionIDEnd)),
			fmt.Sprintf("%s (%s)", "自定义", t.tui.Format(current)),
		}
	}
	sel, err := t.tui.Control().Select(ctx, title, options)
	if err != nil {
		return 0, err
	}
	switch sel {
	case 0:
		return define.DimensionIDOverworld, nil
	case 1:
		return define.DimensionIDNether, nil
	case 2:
		return define.DimensionIDEnd, nil
	case 3:
		if hasCurrent {
			return t.tui.PromptDimension(ctx, fmt.Sprintf("请输入构建维度(当前 %s)：", t.tui.Format(current)), current)
		}
		return t.tui.PromptDimension(ctx, "请输入构建维度：")
	default:
		return 0, fmt.Errorf("未知维度选项：%d", sel)
	}
}

// CurrentDimensionMarker 标记维度菜单中的当前选项。
func CurrentDimensionMarker(current define.Dimension, target define.Dimension) string {
	if current == target {
		return "当前"
	}
	return "-"
}

// fill 命令合并、命令速度、区块组边长

// PromptBuildTaskSpeed 读取构建命令发送速度。
func (t *BuildTaskTUI) PromptBuildTaskSpeed(ctx context.Context, currentValues ...int) (int, error) {
	current := 3000
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.tui.PromptInt(ctx, fmt.Sprintf("请输入命令速度(当前 %d)：", current), current)
}

// PromptBuildTaskChunkGroupSide 读取区块组边长。
func (t *BuildTaskTUI) PromptBuildTaskChunkGroupSide(ctx context.Context, currentValues ...int) (int, error) {
	current := 2
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.tui.PromptInt(ctx, fmt.Sprintf("请输入区块组边长(当前 %d)：", current), current)
}

// PromptBuildTaskConsoleWorldPos 读取 NBT 控制台临时坐标。
func (t *BuildTaskTUI) PromptBuildTaskConsoleWorldPos(ctx context.Context, currentValues ...define.BlockPos) (define.BlockPos, error) {
	current := define.BlockPos{-50, 0, -50}
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.tui.PromptBlockPos(ctx, fmt.Sprintf("请输入 NBT 控制台坐标(当前 %s)：", t.tui.Format(current)), current)
}

// 起始进度、直接进入修补模式、自动进入修补模式、修补超时

// PromptBuildTaskProgress 读取构建起始进度。
func (t *BuildTaskTUI) PromptBuildTaskProgress(ctx context.Context, currentValues ...string) (string, error) {
	current := "0"
	if len(currentValues) > 0 && currentValues[0] != "" {
		current = currentValues[0]
	}
	value, err := t.tui.Control().Prompt(ctx, fmt.Sprintf("请输入起始进度(当前 %s，支持百分比)：", current))
	if err != nil {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		value = current
	}
	return value, nil
}

// PromptBuildTaskFixModeTimeout 读取修补模式超时时间。
func (t *BuildTaskTUI) PromptBuildTaskFixModeTimeout(ctx context.Context, currentValues ...float64) (float64, error) {
	current := 10.0
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.tui.PromptFloat(ctx, fmt.Sprintf("请输入修补模式超时秒数(当前 %s)：", t.tui.Format(current)), current)
}

// PromptBuildTaskGameProgressRefreshDelay 读取游戏内进度刷新间隔。
func (t *BuildTaskTUI) PromptBuildTaskGameProgressRefreshDelay(ctx context.Context, currentValues ...float64) (float64, error) {
	current := 0.5
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.tui.PromptFloat(ctx, fmt.Sprintf("请输入游戏内进度刷新间隔秒数(当前 %s)：", t.tui.Format(current)), current)
}
