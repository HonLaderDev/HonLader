// 构建任务配置模块：
//  1. 建筑来源：建筑文件、建筑区域、源维度
//  2. 构建位置：构建起点、目标维度
//  3. 构建策略：fill 命令合并、命令速度、区块组边长
//  4. 构建处理：清理方块与掉落物、deny 与 border
//  5. 方块处理：命令方块 与 NBT、命令升级
//  6. 加载策略：等待区块加载、常加载区域、预加载、预处理
//  8. 游戏规则：禁命令方块运行、游戏内显示进度
//  9. 断点修补：起始进度、修补模式、修补超时
package ui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/HonLaderDev/HonLader/define"
	build "github.com/HonLaderDev/HonLader/frame/task/build"
)

func valueFromIntPtr(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func valueFromFloatPtr(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

func valueFromBlockPosPtr(value *define.BlockPos, fallback define.BlockPos) define.BlockPos {
	if value == nil {
		return fallback
	}
	return *value
}

func (t *TextUI) PrintBuildTaskPresetSummary(preset buildTaskPreset, config build.BuildTaskConfig) {
	t.c.Print("\n配置摘要：\n")
	t.c.Print(fmt.Sprintf("  预设：%s\n", t.Format.String(preset.Name)))
	t.c.Print(fmt.Sprintf("  建筑文件：%s\n", t.Format.String(config.WorldPath)))
	t.c.Print(fmt.Sprintf("  建筑区域：%s ~ %s\n", t.Format.BlockPos(config.WorldStartPos), t.Format.BlockPos(config.WorldEndPos)))
	t.c.Print(fmt.Sprintf("  建筑维度：%s\n", t.Format.Dimension(config.WorldDimension)))
	t.c.Print(fmt.Sprintf("  构建起点：%s\n", t.Format.BlockPos(config.StartPos)))
	t.c.Print(fmt.Sprintf("  构建维度：%s\n", t.Format.Dimension(config.Dimension)))
	t.c.Print("\n")
}

func (t *TextUI) PrintBuildTaskFullSummary(config build.BuildTaskConfig) {
	t.c.Print("\n配置摘要：\n")
	t.c.Print(fmt.Sprintf("  建筑文件：%s\n", t.Format.String(config.WorldPath)))
	t.c.Print(fmt.Sprintf("  建筑区域：%s ~ %s\n", t.Format.BlockPos(config.WorldStartPos), t.Format.BlockPos(config.WorldEndPos)))
	t.c.Print(fmt.Sprintf("  建筑维度：%s\n", t.Format.Dimension(config.WorldDimension)))
	t.c.Print(fmt.Sprintf("  构建起点：%s\n", t.Format.BlockPos(config.StartPos)))
	t.c.Print(fmt.Sprintf("  构建维度：%s\n", t.Format.Dimension(config.Dimension)))
	t.c.Print(fmt.Sprintf("  fill 命令合并 (%s)\n", t.Format.Bool(!config.DisableAutoFillBuildMode)))
	t.c.Print(fmt.Sprintf("  命令速度：%s\n", strconv.Itoa(valueFromIntPtr(config.Speed, 3000))))
	t.c.Print(fmt.Sprintf("  区块组边长：%s\n", strconv.Itoa(valueFromIntPtr(config.ChunkGroupSide, 2))))
	t.c.Print(fmt.Sprintf("  等待区块加载 (%s)\n", t.Format.Bool(!config.DisableAutoWaitChunkLoad)))
	t.c.Print(fmt.Sprintf("  常加载区域 (%s)\n", t.Format.Bool(config.UseTickingArea)))
	t.c.Print(fmt.Sprintf("  预加载下一组 (%s)\n", t.Format.Bool(config.PreWaitNextChunkLoad)))
	t.c.Print(fmt.Sprintf("  预处理下一组数据 (%s)\n", t.Format.Bool(config.PreHandleNextChunkGroup)))
	t.c.Print(fmt.Sprintf("  构建前清理方块 (%s)\n", t.Format.Bool(config.EnableAutoCleanBlock)))
	t.c.Print(fmt.Sprintf("  构建后清理掉落物 (%s)\n", t.Format.Bool(!config.DisableAutoCleanItem)))
	t.c.Print(fmt.Sprintf("  放置 deny 保护层 (%s)\n", t.Format.Bool(config.EnableAutoPlaceDenyBlock)))
	t.c.Print(fmt.Sprintf("  放置 border 边界 (%s)\n", t.Format.Bool(config.EnableAutoPlaceBorderBlock)))
	t.c.Print(fmt.Sprintf("  构建命令方块 (%s)\n", t.Format.Bool(!config.IgnoreCommandBlock)))
	t.c.Print(fmt.Sprintf("  构建其他 NBT 方块 (%s)\n", t.Format.Bool(!config.IgnoreOtherNBTBlock)))
	t.c.Print(fmt.Sprintf("  自动升级旧命令 (%s)\n", t.Format.Bool(!config.DisableAutoUpgradeCommandBlock)))
	t.c.Print(fmt.Sprintf("  NBT 控制台坐标：%s\n", t.Format.BlockPos(valueFromBlockPosPtr(config.ConsoleWorldPos, define.BlockPos{-50, 0, -50}))))
	t.c.Print(fmt.Sprintf("  禁命令方块运行 (%s)\n", t.Format.Bool(!config.DisableAutoCommandBlocksDisabled)))
	t.c.Print(fmt.Sprintf("  游戏内显示进度 (%s)\n", t.Format.Bool(!config.DisableGameProgress)))
	t.c.Print(fmt.Sprintf("  进度条刷新间隔：%ss\n", t.Format.Float(valueFromFloatPtr(config.GameProgressRefreshDelay, 0.5))))
	t.c.Print(fmt.Sprintf("  起始进度：%s\n", t.Format.OptionalString(config.Progress, "0")))
	t.c.Print(fmt.Sprintf("  直接进入修补模式 (%s)\n", t.Format.Bool(config.EnterFixModeDirectly)))
	t.c.Print(fmt.Sprintf("  自动进入修补模式 (%s)\n", t.Format.Bool(!config.DisableAutoEnterFixMode)))
	t.c.Print(fmt.Sprintf("  修补模式超时：%ss\n", t.Format.Float(valueFromFloatPtr(config.FixModeTimeout, 10))))
	t.c.Print("\n")
}

func (t *TextUI) PromptBuildTaskConfigModule(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.promptBuildTaskConfigModuleSelect(ctx)
		if err != nil {
			return err
		}
		switch sel {
		case 1:
			if err := t.PromptBuildTaskWorldConfig(ctx, config); err != nil {
				return err
			}
		case 2:
			if err := t.PromptBuildTaskPositionConfig(ctx, config); err != nil {
				return err
			}
		case 3:
			if err := t.PromptBuildTaskStrategyConfig(ctx, config); err != nil {
				return err
			}
		case 4:
			if err := t.PromptBuildTaskBuildHandlingConfig(ctx, config); err != nil {
				return err
			}
		case 5:
			if err := t.PromptBuildTaskSpecialBlockConfig(ctx, config); err != nil {
				return err
			}
		case 6:
			if err := t.PromptBuildTaskLoadConfig(ctx, config); err != nil {
				return err
			}
		case 8:
			if err := t.PromptBuildTaskGameRuleConfig(ctx, config); err != nil {
				return err
			}
		case 9:
			if err := t.PromptBuildTaskFixConfig(ctx, config); err != nil {
				return err
			}
		case 10:
			return nil
		}
	}
}

func (t *TextUI) promptBuildTaskConfigModuleSelect(ctx context.Context) (int, error) {
	t.c.Print("请选择配置模块：\n")
	t.c.Print("  [1] 建筑来源：建筑文件、建筑区域、源维度\n")
	t.c.Print("  [2] 构建位置：构建起点、目标维度\n")
	t.c.Print("  [3] 构建策略：fill 命令合并、命令速度、区块组边长\n")
	t.c.Print("  [4] 构建处理：清理方块与掉落物、deny 与 border\n")
	t.c.Print("  [5] 方块处理：命令方块 与 NBT、命令升级\n")
	t.c.Print("  [6] 加载策略：等待区块加载、常加载区域、预加载\n")
	t.c.Print("  [8] 游戏规则：禁命令方块运行、游戏内显示进度\n")
	t.c.Print("  [9] 断点修补：起始进度、修补模式、修补超时\n")
	t.c.Print("  [10] 返回\n")

	for {
		sel, err := t.PromptInt(ctx, "> ")
		if err != nil {
			return 0, err
		}
		switch sel {
		case 1, 2, 3, 4, 5, 6, 8, 9, 10:
			return sel, nil
		default:
			t.c.Print("无效输入，请输入 1~6、8、9、10\n")
		}
	}
}

// 建筑来源：建筑文件、建筑区域、源维度

func (t *TextUI) PromptBuildTaskWorldConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.c.Select(ctx, "请选择建筑来源配置项：\n", []string{
			fmt.Sprintf("%s (%s)", "建筑文件", t.Format.String(config.WorldPath)),
			fmt.Sprintf("%s (%s)", "建筑区域起点", t.Format.BlockPos(config.WorldStartPos)),
			fmt.Sprintf("%s (%s)", "建筑区域终点", t.Format.BlockPos(config.WorldEndPos)),
			fmt.Sprintf("%s (%s)", "建筑区域维度", t.Format.Dimension(config.WorldDimension)),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			config.WorldPath, err = t.PromptBuildTaskWorldPath(ctx, config.WorldPath)
		case 1:
			config.WorldStartPos, err = t.PromptBuildTaskWorldStartPos(ctx, config.WorldStartPos)
		case 2:
			config.WorldEndPos, err = t.PromptBuildTaskWorldEndPos(ctx, config.WorldEndPos)
		case 3:
			config.WorldDimension, err = t.PromptBuildTaskWorldDimension(ctx, "请选择建筑区域维度：\n", config.WorldDimension)
		case 4:
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// 构建位置：构建起点、目标维度

func (t *TextUI) PromptBuildTaskPositionConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.c.Select(ctx, "请选择构建位置配置项：\n", []string{
			fmt.Sprintf("%s (%s)", "构建起点", t.Format.BlockPos(config.StartPos)),
			fmt.Sprintf("%s (%s)", "构建维度", t.Format.Dimension(config.Dimension)),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			config.StartPos, err = t.PromptBuildTaskStartPos(ctx, config.StartPos)
		case 1:
			config.Dimension, err = t.PromptBuildTaskDimension(ctx, "请选择构建维度：\n", config.Dimension)
		case 2:
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// 构建策略：fill 命令合并、命令速度、区块组边长

func (t *TextUI) PromptBuildTaskStrategyConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.c.Select(ctx, "请选择构建策略配置项：\n", []string{
			fmt.Sprintf("%s (%s)", "fill 命令合并", t.Format.Bool(!config.DisableAutoFillBuildMode)),
			fmt.Sprintf("%s (%s)", "命令速度", strconv.Itoa(valueFromIntPtr(config.Speed, 3000))),
			fmt.Sprintf("%s (%s)", "区块组边长", strconv.Itoa(valueFromIntPtr(config.ChunkGroupSide, 2))),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			config.DisableAutoFillBuildMode = !config.DisableAutoFillBuildMode
		case 1:
			speed, err := t.PromptBuildTaskSpeed(ctx, valueFromIntPtr(config.Speed, 3000))
			if err != nil {
				return err
			}
			config.Speed = &speed
		case 2:
			chunkGroupSide, err := t.PromptBuildTaskChunkGroupSide(ctx, valueFromIntPtr(config.ChunkGroupSide, 2))
			if err != nil {
				return err
			}
			config.ChunkGroupSide = &chunkGroupSide
		case 3:
			return nil
		}
	}
}

// 加载策略：等待区块加载、常加载区域、预加载、预处理

func (t *TextUI) PromptBuildTaskLoadConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.c.Select(ctx, "请选择加载策略配置项：\n", []string{
			fmt.Sprintf("%s (%s)", "等待区块加载", t.Format.Bool(!config.DisableAutoWaitChunkLoad)),
			fmt.Sprintf("%s (%s)", "常加载区域", t.Format.Bool(config.UseTickingArea)),
			fmt.Sprintf("%s (%s)", "预加载下一组", t.Format.Bool(config.PreWaitNextChunkLoad)),
			fmt.Sprintf("%s (%s)", "预处理下一组数据", t.Format.Bool(config.PreHandleNextChunkGroup)),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			config.DisableAutoWaitChunkLoad = !config.DisableAutoWaitChunkLoad
		case 1:
			config.UseTickingArea = !config.UseTickingArea
		case 2:
			config.PreWaitNextChunkLoad = !config.PreWaitNextChunkLoad
		case 3:
			config.PreHandleNextChunkGroup = !config.PreHandleNextChunkGroup
		case 4:
			return nil
		}
	}
}

func (t *TextUI) PromptBuildTaskBuildHandlingConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.c.Select(ctx, "请选择构建处理配置项：\n", []string{
			fmt.Sprintf("%s (%s)", "构建前清理方块", t.Format.Bool(config.EnableAutoCleanBlock)),
			fmt.Sprintf("%s (%s)", "构建后清理掉落物", t.Format.Bool(!config.DisableAutoCleanItem)),
			fmt.Sprintf("%s (%s)", "放置 deny 保护层", t.Format.Bool(config.EnableAutoPlaceDenyBlock)),
			fmt.Sprintf("%s (%s)", "放置 border 边界", t.Format.Bool(config.EnableAutoPlaceBorderBlock)),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			config.EnableAutoCleanBlock = !config.EnableAutoCleanBlock
		case 1:
			config.DisableAutoCleanItem = !config.DisableAutoCleanItem
		case 2:
			config.EnableAutoPlaceDenyBlock = !config.EnableAutoPlaceDenyBlock
		case 3:
			config.EnableAutoPlaceBorderBlock = !config.EnableAutoPlaceBorderBlock
		case 4:
			return nil
		}
	}
}

// 特殊方块：命令方块、其他 NBT 方块、旧命令升级、NBT 控制台坐标

func (t *TextUI) PromptBuildTaskSpecialBlockConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.c.Select(ctx, "请选择方块处理配置项：\n", []string{
			fmt.Sprintf("%s (%s)", "构建命令方块", t.Format.Bool(!config.IgnoreCommandBlock)),
			fmt.Sprintf("%s (%s)", "构建其他 NBT 方块", t.Format.Bool(!config.IgnoreOtherNBTBlock)),
			fmt.Sprintf("%s (%s)", "自动升级旧命令", t.Format.Bool(!config.DisableAutoUpgradeCommandBlock)),
			fmt.Sprintf("%s (%s)", "NBT 控制台坐标", t.Format.BlockPos(valueFromBlockPosPtr(config.ConsoleWorldPos, define.BlockPos{-50, 0, -50}))),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			config.IgnoreCommandBlock = !config.IgnoreCommandBlock
		case 1:
			config.IgnoreOtherNBTBlock = !config.IgnoreOtherNBTBlock
		case 2:
			config.DisableAutoUpgradeCommandBlock = !config.DisableAutoUpgradeCommandBlock
		case 3:
			consoleWorldPos, err := t.PromptBuildTaskConsoleWorldPos(ctx, valueFromBlockPosPtr(config.ConsoleWorldPos, define.BlockPos{-50, 0, -50}))
			if err != nil {
				return err
			}
			config.ConsoleWorldPos = &consoleWorldPos
		case 4:
			return nil
		}
	}
}

// 断点与修补：起始进度、直接进入修补模式、自动进入修补模式、修补超时

func (t *TextUI) PromptBuildTaskFixConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.c.Select(ctx, "请选择断点修补配置项：\n", []string{
			fmt.Sprintf("%s (%s)", "起始进度", t.Format.OptionalString(config.Progress, "0")),
			fmt.Sprintf("%s (%s)", "直接进入修补模式", t.Format.Bool(config.EnterFixModeDirectly)),
			fmt.Sprintf("%s (%s)", "自动进入修补模式", t.Format.Bool(!config.DisableAutoEnterFixMode)),
			fmt.Sprintf("%s (%s)", "修补模式超时", t.Format.Float(valueFromFloatPtr(config.FixModeTimeout, 10))+"s"),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			progress, err := t.PromptBuildTaskProgress(ctx, t.Format.OptionalString(config.Progress, "0"))
			if err != nil {
				return err
			}
			config.Progress = progress
		case 1:
			config.EnterFixModeDirectly = !config.EnterFixModeDirectly
		case 2:
			config.DisableAutoEnterFixMode = !config.DisableAutoEnterFixMode
		case 3:
			fixModeTimeout, err := t.PromptBuildTaskFixModeTimeout(ctx, valueFromFloatPtr(config.FixModeTimeout, 10))
			if err != nil {
				return err
			}
			config.FixModeTimeout = &fixModeTimeout
		case 4:
			return nil
		}
	}
}

func (t *TextUI) PromptBuildTaskGameRuleConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.c.Select(ctx, "请选择游戏规则配置项：\n", []string{
			fmt.Sprintf("%s (%s)", "禁命令方块运行", t.Format.Bool(!config.DisableAutoCommandBlocksDisabled)),
			fmt.Sprintf("%s (%s)", "游戏内显示进度", t.Format.Bool(!config.DisableGameProgress)),
			fmt.Sprintf("%s (%s)", "进度条刷新间隔", t.Format.Float(valueFromFloatPtr(config.GameProgressRefreshDelay, 0.5))+"s"),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			config.DisableAutoCommandBlocksDisabled = !config.DisableAutoCommandBlocksDisabled
		case 1:
			config.DisableGameProgress = !config.DisableGameProgress
		case 2:
			gameProgressRefreshDelay, err := t.PromptBuildTaskGameProgressRefreshDelay(ctx, valueFromFloatPtr(config.GameProgressRefreshDelay, 0.5))
			if err != nil {
				return err
			}
			config.GameProgressRefreshDelay = &gameProgressRefreshDelay
		case 3:
			return nil
		}
	}
}

// 建筑文件、建筑区域、源维度

func (t *TextUI) PromptBuildTaskWorldPath(ctx context.Context, currentValues ...string) (string, error) {
	if len(currentValues) > 0 && currentValues[0] != "" {
		return t.PromptRequired(ctx, fmt.Sprintf("请输入建筑文件路径(当前 %s)：", currentValues[0]), currentValues[0])
	}
	return t.PromptRequired(ctx, "请输入建筑文件路径：")
}

func (t *TextUI) PromptBuildTaskWorldStartPos(ctx context.Context, currentValues ...define.BlockPos) (define.BlockPos, error) {
	if len(currentValues) > 0 {
		return t.PromptBlockPos(ctx, fmt.Sprintf("请输入建筑区域起点(当前 %s)：", t.Format.BlockPos(currentValues[0])), currentValues[0])
	}
	return t.PromptBlockPos(ctx, "请输入建筑区域起点(x,y,z)：")
}

func (t *TextUI) PromptBuildTaskWorldEndPos(ctx context.Context, currentValues ...define.BlockPos) (define.BlockPos, error) {
	if len(currentValues) > 0 {
		return t.PromptBlockPos(ctx, fmt.Sprintf("请输入建筑区域终点(当前 %s)：", t.Format.BlockPos(currentValues[0])), currentValues[0])
	}
	return t.PromptBlockPos(ctx, "请输入建筑区域终点(x,y,z)：")
}

func (t *TextUI) PromptBuildTaskWorldDimension(ctx context.Context, args ...any) (define.Dimension, error) {
	title, current, hasCurrent := buildTaskDimensionPromptArgs("请选择建筑区域维度：\n", args...)
	return t.PromptBuildTaskDimensionMenu(ctx, title, current, hasCurrent)
}

// 构建起点、目标维度

func (t *TextUI) PromptBuildTaskStartPos(ctx context.Context, currentValues ...define.BlockPos) (define.BlockPos, error) {
	if len(currentValues) > 0 {
		return t.PromptBlockPos(ctx, fmt.Sprintf("请输入构建起点(当前 %s)：", t.Format.BlockPos(currentValues[0])), currentValues[0])
	}
	return t.PromptBlockPos(ctx, "请输入构建起点(x,y,z)：")
}

func (t *TextUI) PromptBuildTaskDimension(ctx context.Context, args ...any) (define.Dimension, error) {
	title, current, hasCurrent := buildTaskDimensionPromptArgs("请选择构建维度：\n", args...)
	return t.PromptBuildTaskDimensionMenu(ctx, title, current, hasCurrent)
}

func buildTaskDimensionPromptArgs(defaultTitle string, args ...any) (string, define.Dimension, bool) {
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

func (t *TextUI) PromptBuildTaskDimensionMenu(ctx context.Context, title string, current define.Dimension, hasCurrent bool) (define.Dimension, error) {
	options := []string{"主世界", "下界", "末地", "自定义"}
	if hasCurrent {
		options = []string{
			fmt.Sprintf("%s (%s)", "主世界", currentDimensionMarker(current, define.DimensionIDOverworld)),
			fmt.Sprintf("%s (%s)", "下界", currentDimensionMarker(current, define.DimensionIDNether)),
			fmt.Sprintf("%s (%s)", "末地", currentDimensionMarker(current, define.DimensionIDEnd)),
			fmt.Sprintf("%s (%s)", "自定义", t.Format.Dimension(current)),
		}
	}
	sel, err := t.c.Select(ctx, title, options)
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
			return t.PromptDimension(ctx, fmt.Sprintf("请输入构建维度(当前 %s)：", t.Format.Dimension(current)), current)
		}
		return t.PromptDimension(ctx, "请输入构建维度：")
	default:
		return 0, fmt.Errorf("未知维度选项：%d", sel)
	}
}

func currentDimensionMarker(current define.Dimension, target define.Dimension) string {
	if current == target {
		return "当前"
	}
	return "-"
}

// fill 命令合并、命令速度、区块组边长

func (t *TextUI) PromptBuildTaskSpeed(ctx context.Context, currentValues ...int) (int, error) {
	current := 3000
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.PromptInt(ctx, fmt.Sprintf("请输入命令速度(当前 %d)：", current), current)
}

func (t *TextUI) PromptBuildTaskChunkGroupSide(ctx context.Context, currentValues ...int) (int, error) {
	current := 2
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.PromptInt(ctx, fmt.Sprintf("请输入区块组边长(当前 %d)：", current), current)
}

func (t *TextUI) PromptBuildTaskConsoleWorldPos(ctx context.Context, currentValues ...define.BlockPos) (define.BlockPos, error) {
	current := define.BlockPos{-50, 0, -50}
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.PromptBlockPos(ctx, fmt.Sprintf("请输入 NBT 控制台坐标(当前 %s)：", t.Format.BlockPos(current)), current)
}

// 起始进度、直接进入修补模式、自动进入修补模式、修补超时

func (t *TextUI) PromptBuildTaskProgress(ctx context.Context, currentValues ...string) (string, error) {
	current := "0"
	if len(currentValues) > 0 && currentValues[0] != "" {
		current = currentValues[0]
	}
	value, err := t.c.Prompt(ctx, fmt.Sprintf("请输入起始进度(当前 %s，支持百分比)：", current))
	if err != nil {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		value = current
	}
	return value, nil
}

func (t *TextUI) PromptBuildTaskFixModeTimeout(ctx context.Context, currentValues ...float64) (float64, error) {
	current := 10.0
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.PromptFloat(ctx, fmt.Sprintf("请输入修补模式超时秒数(当前 %s)：", t.Format.Float(current)), current)
}

func (t *TextUI) PromptBuildTaskGameProgressRefreshDelay(ctx context.Context, currentValues ...float64) (float64, error) {
	current := 0.5
	if len(currentValues) > 0 {
		current = currentValues[0]
	}
	return t.PromptFloat(ctx, fmt.Sprintf("请输入游戏内进度刷新间隔秒数(当前 %s)：", t.Format.Float(current)), current)
}
