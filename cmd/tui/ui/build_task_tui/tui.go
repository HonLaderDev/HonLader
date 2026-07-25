package build_task_tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tui_define "github.com/HonLaderDev/HonLader/cmd/tui/define"
	"github.com/HonLaderDev/HonLader/cmd/tui/ui/structure_convert_tui"
	"github.com/HonLaderDev/HonLader/define"
	build "github.com/HonLaderDev/HonLader/frame/task/build"
	"github.com/HonLaderDev/HonLader/utils"
	structure_utils "github.com/HonLaderDev/HonLader/utils/structure"
)

// BuildTaskTUI 负责构建任务配置相关的文本交互。
type BuildTaskTUI struct {
	tui tui_define.TextUI
}

// NewBuildTaskTUI 创建构建任务配置交互实例。
func NewBuildTaskTUI(tui tui_define.TextUI) *BuildTaskTUI {
	return &BuildTaskTUI{tui: tui}
}

// CreateBuildTaskConfig 通过快速向导和可选高级配置创建构建任务配置。
func (t *BuildTaskTUI) CreateBuildTaskConfig(ctx context.Context) (build.BuildTaskConfig, error) {
	config := build.BuildTaskConfig{}
	config.FillDefault()
	if err := t.PromptBuildTaskConfigTemplate(ctx, &config); err != nil {
		return build.BuildTaskConfig{}, err
	}

	worldPath, err := t.PromptBuildTaskWorldPath(ctx)
	if err != nil {
		return build.BuildTaskConfig{}, err
	}
	worldPath, err = t.EnsureBuildTaskWorldPath(ctx, worldPath)
	if err != nil {
		return build.BuildTaskConfig{}, err
	}
	worldStartPos, worldEndPos, parsedWorldRange, err := t.PromptBuildTaskWorldRange(ctx, worldPath)
	if err != nil {
		return build.BuildTaskConfig{}, err
	}
	worldDimension := define.Dimension(define.DimensionIDOverworld)
	if parsedWorldRange {
		t.tui.Control().Print("已从建筑文件名解析建筑区域，建筑维度使用主世界\n")
	} else {
		worldDimension, err = t.PromptBuildTaskWorldDimension(ctx, "请选择建筑区域维度：\n")
		if err != nil {
			return build.BuildTaskConfig{}, err
		}
	}
	startPos, err := t.PromptBuildTaskStartPos(ctx)
	if err != nil {
		return build.BuildTaskConfig{}, err
	}
	dimension, err := t.PromptBuildTaskDimension(ctx, "请选择构建维度：\n")
	if err != nil {
		return build.BuildTaskConfig{}, err
	}

	config.WorldPath = worldPath
	config.WorldStartPos = worldStartPos
	config.WorldEndPos = worldEndPos
	config.WorldDimension = worldDimension
	config.StartPos = startPos
	config.Dimension = dimension

	t.PrintBuildTaskSummary(config)
	continueConfig, err := t.tui.Control().Confirm(ctx, "是否继续配置[y/N]：", false)
	if err != nil {
		return build.BuildTaskConfig{}, err
	}
	if continueConfig {
		if err := t.PromptBuildTaskConfigModule(ctx, &config); err != nil {
			return build.BuildTaskConfig{}, err
		}
		t.PrintBuildTaskSummary(config)
	}

	confirm, err := t.tui.Control().Confirm(ctx, "是否进行构建[Y/n]：", true)
	if err != nil {
		return build.BuildTaskConfig{}, err
	}
	if confirm {
		return config, nil
	}
	return build.BuildTaskConfig{}, fmt.Errorf("已取消构建")
}

// EnsureBuildTaskWorldPath 确保建筑文件路径可作为 mcworld 使用。
func (t *BuildTaskTUI) EnsureBuildTaskWorldPath(ctx context.Context, worldPath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	info, err := os.Stat(worldPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return worldPath, nil
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("建筑文件路径不是普通文件或目录：%s", worldPath)
	}
	if IsMCWorldPath(worldPath) {
		return worldPath, nil
	}
	defaultPath := DefaultConvertedMCWorldPath(worldPath)
	if _, err := os.Stat(defaultPath); err == nil {
		t.tui.Control().Print(fmt.Sprintf("检测到同名 mcworld 文件，直接使用：%s\n", defaultPath))
		return defaultPath, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	destPath, worldName, err := t.PromptBuildTaskConvertOutput(ctx, worldPath, defaultPath)
	if err != nil {
		return "", err
	}
	converter := structure_utils.NewStructureConverter()
	convertTUI := structure_convert_tui.NewStructureConvertTUI(t.tui)
	convertTUI.Register(converter)
	convertTUI.Reset()
	if _, err := converter.ConvertToMCWorldWithName(worldPath, destPath, worldName); err != nil {
		convertTUI.Reset()
		_ = os.Remove(destPath)
		return "", fmt.Errorf("结构转换为 mcworld 失败：%w", err)
	}
	return destPath, nil
}

// PromptBuildTaskConvertOutput 读取结构转换输出路径或名称。
func (t *BuildTaskTUI) PromptBuildTaskConvertOutput(ctx context.Context, srcPath, defaultPath string) (string, string, error) {
	input, err := t.tui.PromptRequired(ctx, fmt.Sprintf("请输入转换后的 mcworld 路径或名称(当前 %s)：", defaultPath), defaultPath)
	if err != nil {
		return "", "", err
	}
	destPath := NormalizeConvertedMCWorldPath(srcPath, input)
	worldName := structure_utils.DefaultWorldName(destPath)
	return destPath, worldName, nil
}

// IsMCWorldPath 判断路径是否是 mcworld 文件。
func IsMCWorldPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".mcworld")
}

// DefaultConvertedMCWorldPath 生成默认转换输出路径。
func DefaultConvertedMCWorldPath(srcPath string) string {
	name := structure_utils.DefaultWorldName(srcPath)
	return filepath.Join(filepath.Dir(srcPath), name+".mcworld")
}

// NormalizeConvertedMCWorldPath 规范化用户输入的转换输出路径。
func NormalizeConvertedMCWorldPath(srcPath, input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return DefaultConvertedMCWorldPath(srcPath)
	}
	if filepath.Dir(input) == "." {
		input = filepath.Join(filepath.Dir(srcPath), input)
	}
	if !strings.EqualFold(filepath.Ext(input), ".mcworld") {
		input += ".mcworld"
	}
	return input
}

// PromptBuildTaskWorldRange 读取建筑区域，优先复用建筑文件名里的范围。
func (t *BuildTaskTUI) PromptBuildTaskWorldRange(ctx context.Context, worldPath string) (define.BlockPos, define.BlockPos, bool, error) {
	start, end, found := utils.ParseWorldRange(worldPath)
	if found {
		t.tui.Control().Print(fmt.Sprintf("已从建筑文件名解析建筑区域：%s ~ %s\n", t.tui.Format(start), t.tui.Format(end)))
		return start, end, true, nil
	}
	start, end, found, err := utils.ParseWorldRangeFromZipLevelName(worldPath)
	if err == nil && found {
		t.tui.Control().Print(fmt.Sprintf("已从建筑文件 levelname.txt 解析建筑区域：%s ~ %s\n", t.tui.Format(start), t.tui.Format(end)))
		return start, end, true, nil
	}

	worldStartPos, err := t.PromptBuildTaskWorldStartPos(ctx)
	if err != nil {
		return define.BlockPos{}, define.BlockPos{}, false, err
	}
	worldEndPos, err := t.PromptBuildTaskWorldEndPos(ctx)
	if err != nil {
		return define.BlockPos{}, define.BlockPos{}, false, err
	}
	return worldStartPos, worldEndPos, false, nil
}

// PromptBuildTaskConfigTemplate 询问并应用构建任务配置模版。
func (t *BuildTaskTUI) PromptBuildTaskConfigTemplate(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return fmt.Errorf("构建任务配置不能为空")
	}
	sel, err := t.tui.Control().Select(ctx, "请选择配置模版：\n", []string{
		"标准构建，构建所有方块",
		"纯建筑式，构建普通方块",
		"修补模式，进行修补建筑",
	})
	if err != nil {
		return err
	}
	switch sel {
	case 0:
		config.IgnoreCommandBlock = false
		config.IgnoreOtherNBTBlock = false
		config.EnterFixModeDirectly = false
	case 1:
		config.IgnoreCommandBlock = true
		config.IgnoreOtherNBTBlock = true
		config.EnterFixModeDirectly = false
	case 2:
		return t.PromptBuildTaskFixTemplate(ctx, config)
	default:
		return fmt.Errorf("未知配置模版：%d", sel)
	}
	return nil
}

// PromptBuildTaskFixTemplate 询问并应用修补模式的启动方式。
func (t *BuildTaskTUI) PromptBuildTaskFixTemplate(ctx context.Context, config *build.BuildTaskConfig) error {
	sel, err := t.tui.Control().Select(ctx, "请选择修补方式：\n", []string{
		"立即进入修补模式",
		"从指定进度开始，构建完成后自动进入修补模式",
	})
	if err != nil {
		return err
	}
	switch sel {
	case 0:
		config.EnterFixModeDirectly = true
	case 1:
		progress, err := t.PromptBuildTaskProgress(ctx, config.Progress)
		if err != nil {
			return err
		}
		config.Progress = progress
		config.EnterFixModeDirectly = false
		config.DisableAutoEnterFixMode = false
	default:
		return fmt.Errorf("未知修补方式：%d", sel)
	}
	return nil
}
