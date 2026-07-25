// 构建任务配置模块：
//  1. 建筑来源：建筑文件、建筑区域、源维度
//  2. 构建位置：构建起点、目标维度
//  3. 构建策略：fill 命令合并、命令速度、区块组边长
//  4. 构建处理：清理方块与掉落物、deny 与 border
//  5. 方块处理：命令方块 与 NBT、命令升级
//  6. 加载策略：等待区块加载、常加载区域、预加载、预处理
//  8. 游戏规则：禁命令方块运行、游戏内显示进度
//  9. 断点修补：起始进度、修补模式、修补超时
package build_task_tui

import (
	"context"
	"errors"
	"fmt"

	"github.com/HonLaderDev/HonLader/define"
	build "github.com/HonLaderDev/HonLader/frame/task/build"
	"github.com/HonLaderDev/HonLader/utils"
)

// ConfigOption 格式化带当前值的配置菜单项。
func (t *BuildTaskTUI) ConfigOption(name string, value any) string {
	return fmt.Sprintf("%s (%s)", name, t.tui.Format(value))
}

// PromptBuildTaskConfigModule 循环展示高级配置模块菜单并应用用户修改。
func (t *BuildTaskTUI) PromptBuildTaskConfigModule(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		t.tui.Control().Print("请选择配置模块：\n")
		t.tui.Control().Print("  [1] 建筑来源：建筑文件、建筑区域、源维度\n")
		t.tui.Control().Print("  [2] 构建位置：构建起点、目标维度\n")
		t.tui.Control().Print("  [3] 构建策略：fill 命令合并、命令速度、区块组边长\n")
		t.tui.Control().Print("  [4] 构建处理：清理方块与掉落物、deny 与 border\n")
		t.tui.Control().Print("  [5] 方块处理：命令方块 与 NBT、命令升级\n")
		t.tui.Control().Print("  [6] 加载策略：等待区块加载、常加载区域、预加载\n")
		t.tui.Control().Print("  [8] 游戏规则：禁命令方块运行、游戏内显示进度\n")
		t.tui.Control().Print("  [9] 断点修补：起始进度、修补模式、修补超时\n")
		t.tui.Control().Print("  [10] 返回\n")

		sel, err := t.tui.PromptInt(ctx, "> ")
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
		default:
			t.tui.Control().Print("无效输入，请输入 1~6、8、9、10\n")
		}
	}
}

// 建筑来源：建筑文件、建筑区域、源维度

// PromptBuildTaskWorldConfig 交互式修改建筑来源相关配置。
func (t *BuildTaskTUI) PromptBuildTaskWorldConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.tui.Control().Select(ctx, "请选择建筑来源配置项：\n", []string{
			t.ConfigOption("建筑文件", config.WorldPath),
			t.ConfigOption("建筑区域起点", config.WorldStartPos),
			t.ConfigOption("建筑区域终点", config.WorldEndPos),
			t.ConfigOption("建筑区域维度", config.WorldDimension),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			config.WorldPath, err = t.PromptBuildTaskWorldPath(ctx, config.WorldPath)
			if err == nil {
				config.WorldPath, err = t.EnsureBuildTaskWorldPath(ctx, config.WorldPath)
			}
			if err == nil {
				t.ApplyWorldRangeFromPath(config)
			}
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

// ApplyWorldRangeFromPath 从建筑文件路径中解析并更新建筑区域。
func (t *BuildTaskTUI) ApplyWorldRangeFromPath(config *build.BuildTaskConfig) {
	if config == nil {
		return
	}
	start, end, found := utils.ParseWorldRange(config.WorldPath)
	source := "建筑文件名"
	if !found {
		var err error
		start, end, found, err = utils.ParseWorldRangeFromZipLevelName(config.WorldPath)
		if err != nil || !found {
			return
		}
		source = "建筑文件 levelname.txt"
	}
	config.WorldStartPos = start
	config.WorldEndPos = end
	config.WorldDimension = define.DimensionIDOverworld
	t.tui.Control().Print(fmt.Sprintf("已从%s解析建筑区域：%s ~ %s，建筑维度使用主世界\n", source, t.tui.Format(start), t.tui.Format(end)))
}

// 构建位置：构建起点、目标维度

// PromptBuildTaskPositionConfig 交互式修改构建位置相关配置。
func (t *BuildTaskTUI) PromptBuildTaskPositionConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.tui.Control().Select(ctx, "请选择构建位置配置项：\n", []string{
			t.ConfigOption("构建起点", config.StartPos),
			t.ConfigOption("构建维度", config.Dimension),
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

// PromptBuildTaskStrategyConfig 交互式修改构建策略相关配置。
func (t *BuildTaskTUI) PromptBuildTaskStrategyConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	config.FillDefault()
	for {
		sel, err := t.tui.Control().Select(ctx, "请选择构建策略配置项：\n", []string{
			t.ConfigOption("fill 命令合并", !config.DisableAutoFillBuildMode),
			t.ConfigOption("命令速度", *config.Speed),
			t.ConfigOption("区块组边长", *config.ChunkGroupSide),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			config.DisableAutoFillBuildMode = !config.DisableAutoFillBuildMode
		case 1:
			speed, err := t.PromptBuildTaskSpeed(ctx, *config.Speed)
			if err != nil {
				return err
			}
			config.Speed = &speed
		case 2:
			chunkGroupSide, err := t.PromptBuildTaskChunkGroupSide(ctx, *config.ChunkGroupSide)
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

// PromptBuildTaskLoadConfig 交互式修改区块加载相关配置。
func (t *BuildTaskTUI) PromptBuildTaskLoadConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.tui.Control().Select(ctx, "请选择加载策略配置项：\n", []string{
			t.ConfigOption("等待区块加载", !config.DisableAutoWaitChunkLoad),
			t.ConfigOption("常加载区域", config.UseTickingArea),
			t.ConfigOption("预加载下一组", config.PreWaitNextChunkLoad),
			t.ConfigOption("预处理下一组数据", config.PreHandleNextChunkGroup),
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

// PromptBuildTaskBuildHandlingConfig 交互式修改构建清理和边界处理配置。
func (t *BuildTaskTUI) PromptBuildTaskBuildHandlingConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	for {
		sel, err := t.tui.Control().Select(ctx, "请选择构建处理配置项：\n", []string{
			t.ConfigOption("构建前清理方块", config.EnableAutoCleanBlock),
			t.ConfigOption("构建后清理掉落物", !config.DisableAutoCleanItem),
			t.ConfigOption("放置 deny 保护层", config.EnableAutoPlaceDenyBlock),
			t.ConfigOption("放置 border 边界", config.EnableAutoPlaceBorderBlock),
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

// PromptBuildTaskSpecialBlockConfig 交互式修改特殊方块处理配置。
func (t *BuildTaskTUI) PromptBuildTaskSpecialBlockConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	config.FillDefault()
	for {
		sel, err := t.tui.Control().Select(ctx, "请选择方块处理配置项：\n", []string{
			t.ConfigOption("构建命令方块", !config.IgnoreCommandBlock),
			t.ConfigOption("构建其他 NBT 方块", !config.IgnoreOtherNBTBlock),
			t.ConfigOption("自动升级旧命令", !config.DisableAutoUpgradeCommandBlock),
			t.ConfigOption("NBT 控制台坐标", *config.ConsoleWorldPos),
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
			consoleWorldPos, err := t.PromptBuildTaskConsoleWorldPos(ctx, *config.ConsoleWorldPos)
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

// PromptBuildTaskFixConfig 交互式修改断点和修补模式配置。
func (t *BuildTaskTUI) PromptBuildTaskFixConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	config.FillDefault()
	for {
		sel, err := t.tui.Control().Select(ctx, "请选择断点修补配置项：\n", []string{
			t.ConfigOption("起始进度", config.Progress),
			t.ConfigOption("直接进入修补模式", config.EnterFixModeDirectly),
			t.ConfigOption("自动进入修补模式", !config.DisableAutoEnterFixMode),
			t.ConfigOption("修补模式超时", t.tui.Format(*config.FixModeTimeout)+"s"),
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			progress, err := t.PromptBuildTaskProgress(ctx, t.tui.Format(config.Progress))
			if err != nil {
				return err
			}
			config.Progress = progress
		case 1:
			config.EnterFixModeDirectly = !config.EnterFixModeDirectly
		case 2:
			config.DisableAutoEnterFixMode = !config.DisableAutoEnterFixMode
		case 3:
			fixModeTimeout, err := t.PromptBuildTaskFixModeTimeout(ctx, *config.FixModeTimeout)
			if err != nil {
				return err
			}
			config.FixModeTimeout = &fixModeTimeout
		case 4:
			return nil
		}
	}
}

// PromptBuildTaskGameRuleConfig 交互式修改游戏规则和进度显示配置。
func (t *BuildTaskTUI) PromptBuildTaskGameRuleConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return errors.New("构建任务配置不能为空")
	}
	config.FillDefault()
	for {
		sel, err := t.tui.Control().Select(ctx, "请选择游戏规则配置项：\n", []string{
			t.ConfigOption("禁命令方块运行", !config.DisableAutoCommandBlocksDisabled),
			t.ConfigOption("游戏内显示进度", !config.DisableGameProgress),
			t.ConfigOption("进度条刷新间隔", t.tui.Format(*config.GameProgressRefreshDelay)+"s"),
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
			gameProgressRefreshDelay, err := t.PromptBuildTaskGameProgressRefreshDelay(ctx, *config.GameProgressRefreshDelay)
			if err != nil {
				return err
			}
			config.GameProgressRefreshDelay = &gameProgressRefreshDelay
		case 3:
			return nil
		}
	}
}
