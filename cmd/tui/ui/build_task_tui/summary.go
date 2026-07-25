package build_task_tui

import (
	"fmt"

	build "github.com/HonLaderDev/HonLader/frame/task/build"
)

// PrintBuildTaskSummary 打印构建任务配置摘要。
func (t *BuildTaskTUI) PrintBuildTaskSummary(config build.BuildTaskConfig) {
	config.FillDefault()
	t.tui.Control().Print("\n配置摘要：\n")
	t.tui.Control().Print(fmt.Sprintf("  建筑文件：%s\n", t.tui.Format(config.WorldPath)))
	t.tui.Control().Print(fmt.Sprintf("  建筑区域：%s ~ %s\n", t.tui.Format(config.WorldStartPos), t.tui.Format(config.WorldEndPos)))
	t.tui.Control().Print(fmt.Sprintf("  建筑维度：%s\n", t.tui.Format(config.WorldDimension)))
	t.tui.Control().Print(fmt.Sprintf("  构建起点：%s\n", t.tui.Format(config.StartPos)))
	t.tui.Control().Print(fmt.Sprintf("  构建维度：%s\n", t.tui.Format(config.Dimension)))
	t.tui.Control().Print(fmt.Sprintf("  fill 命令合并 (%s)\n", t.tui.Format(!config.DisableAutoFillBuildMode)))
	t.tui.Control().Print(fmt.Sprintf("  命令速度：%s\n", t.tui.Format(*config.Speed)))
	t.tui.Control().Print(fmt.Sprintf("  区块组边长：%s\n", t.tui.Format(*config.ChunkGroupSide)))
	t.tui.Control().Print(fmt.Sprintf("  等待区块加载 (%s)\n", t.tui.Format(!config.DisableAutoWaitChunkLoad)))
	t.tui.Control().Print(fmt.Sprintf("  常加载区域 (%s)\n", t.tui.Format(config.UseTickingArea)))
	t.tui.Control().Print(fmt.Sprintf("  预加载下一组 (%s)\n", t.tui.Format(config.PreWaitNextChunkLoad)))
	t.tui.Control().Print(fmt.Sprintf("  预处理下一组数据 (%s)\n", t.tui.Format(config.PreHandleNextChunkGroup)))
	t.tui.Control().Print(fmt.Sprintf("  构建前清理方块 (%s)\n", t.tui.Format(config.EnableAutoCleanBlock)))
	t.tui.Control().Print(fmt.Sprintf("  构建后清理掉落物 (%s)\n", t.tui.Format(!config.DisableAutoCleanItem)))
	t.tui.Control().Print(fmt.Sprintf("  放置 deny 保护层 (%s)\n", t.tui.Format(config.EnableAutoPlaceDenyBlock)))
	t.tui.Control().Print(fmt.Sprintf("  放置 border 边界 (%s)\n", t.tui.Format(config.EnableAutoPlaceBorderBlock)))
	t.tui.Control().Print(fmt.Sprintf("  构建命令方块 (%s)\n", t.tui.Format(!config.IgnoreCommandBlock)))
	t.tui.Control().Print(fmt.Sprintf("  构建其他 NBT 方块 (%s)\n", t.tui.Format(!config.IgnoreOtherNBTBlock)))
	t.tui.Control().Print(fmt.Sprintf("  自动升级旧命令 (%s)\n", t.tui.Format(!config.DisableAutoUpgradeCommandBlock)))
	t.tui.Control().Print(fmt.Sprintf("  NBT 控制台坐标：%s\n", t.tui.Format(*config.ConsoleWorldPos)))
	t.tui.Control().Print(fmt.Sprintf("  禁命令方块运行 (%s)\n", t.tui.Format(!config.DisableAutoCommandBlocksDisabled)))
	t.tui.Control().Print(fmt.Sprintf("  游戏内显示进度 (%s)\n", t.tui.Format(!config.DisableGameProgress)))
	t.tui.Control().Print(fmt.Sprintf("  进度条刷新间隔：%ss\n", t.tui.Format(*config.GameProgressRefreshDelay)))
	t.tui.Control().Print(fmt.Sprintf("  起始进度：%s\n", t.tui.Format(config.Progress)))
	t.tui.Control().Print(fmt.Sprintf("  直接进入修补模式 (%s)\n", t.tui.Format(config.EnterFixModeDirectly)))
	t.tui.Control().Print(fmt.Sprintf("  自动进入修补模式 (%s)\n", t.tui.Format(!config.DisableAutoEnterFixMode)))
	t.tui.Control().Print(fmt.Sprintf("  修补模式超时：%ss\n", t.tui.Format(*config.FixModeTimeout)))
	t.tui.Control().Print("\n")
}
