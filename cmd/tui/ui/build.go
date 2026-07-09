package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/HonLaderDev/HonLader/define"
	build "github.com/HonLaderDev/HonLader/frame/task/build"
)

type buildTaskPreset struct {
	Name string
}

func (t *TextUI) StartBuild(ctx context.Context) error {
	if _, err := t.SelectServerConfig(ctx); err != nil {
		return err
	}

	name := fmt.Sprintf("构建任务 %s", time.Now().Format("2006-01-02 15:04:05"))

	config, err := t.CreateBuildTaskConfig(ctx)
	if err != nil {
		return err
	}

	if err := t.l.Connect(ctx); err != nil {
		return fmt.Errorf("连接 Core 失败：%w", err)
	}

	t.l.WithTaskGroupName(name).AddTask(config.NewTask(t.l))
	return t.l.Start()
}

func (t *TextUI) CreateBuildTaskConfig(ctx context.Context) (build.BuildTaskConfig, error) {
	config := build.BuildTaskConfig{}
	preset, err := t.PromptBuildTaskPreset(ctx, &config)
	if err != nil {
		return build.BuildTaskConfig{}, err
	}
	if err := t.promptBuildTaskRequiredConfig(ctx, &config); err != nil {
		return build.BuildTaskConfig{}, err
	}

	t.PrintBuildTaskPresetSummary(preset, config)
	continueConfig, err := t.c.Confirm(ctx, "是否继续配置[y/N]：", false)
	if err != nil {
		return build.BuildTaskConfig{}, err
	}
	if continueConfig {
		if err := t.PromptBuildTaskConfigModule(ctx, &config); err != nil {
			return build.BuildTaskConfig{}, err
		}
		t.PrintBuildTaskFullSummary(config)
	}

	confirm, err := t.c.Confirm(ctx, "是否进行构建[Y/n]：", true)
	if err != nil {
		return build.BuildTaskConfig{}, err
	}
	if confirm {
		return config, nil
	}
	return build.BuildTaskConfig{}, fmt.Errorf("已取消构建")
}

func (t *TextUI) PromptBuildTaskPreset(ctx context.Context, config *build.BuildTaskConfig) (buildTaskPreset, error) {
	if config == nil {
		return buildTaskPreset{}, fmt.Errorf("构建任务配置不能为空")
	}
	sel, err := t.c.Select(ctx, "请选择构建预设：\n", []string{
		"标准：稳定默认，适合大多数构建时",
		"快速：提高速度，适合服务器较好时",
		"保守：降低速度，适合服务器较卡时",
		"修补：指定进度，进入修补相关流程",
	})
	if err != nil {
		return buildTaskPreset{}, err
	}
	speed := 3000
	chunkGroupSide := 2
	fixModeTimeout := 10.0
	config.Speed = &speed
	config.ChunkGroupSide = &chunkGroupSide
	config.DisableAutoFillBuildMode = false
	config.DisableAutoWaitChunkLoad = false
	config.UseTickingArea = false
	config.PreHandleNextChunkGroup = false
	config.PreWaitNextChunkLoad = false
	config.EnableAutoCleanBlock = false
	config.DisableAutoCleanItem = false
	config.DisableAutoEnterFixMode = false
	config.FixModeTimeout = &fixModeTimeout
	config.ConsoleWorldPos = &define.BlockPos{-50, 0, -50}
	config.GameProgressRefreshDelay = floatPtr(0.5)

	switch sel {
	case 0:
		return buildTaskPreset{Name: "标准"}, nil
	case 1:
		speed = 6000
		chunkGroupSide = 3
		config.UseTickingArea = true
		config.PreHandleNextChunkGroup = true
		config.PreWaitNextChunkLoad = true
		return buildTaskPreset{Name: "快速"}, nil
	case 2:
		speed = 1200
		chunkGroupSide = 1
		return buildTaskPreset{Name: "保守"}, nil
	case 3:
		if err := t.promptBuildTaskPatchPreset(ctx, config); err != nil {
			return buildTaskPreset{}, err
		}
		return buildTaskPreset{Name: "修补"}, nil
	default:
		return buildTaskPreset{}, fmt.Errorf("未知构建预设：%d", sel)
	}
}

func (t *TextUI) promptBuildTaskPatchPreset(ctx context.Context, config *build.BuildTaskConfig) error {
	sel, err := t.c.Select(ctx, "请选择修补方式：\n", []string{
		"从指定进度继续构建",
		"直接进入修补模式",
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
		config.EnterFixModeDirectly = true
	}
	return nil
}

func floatPtr(value float64) *float64 {
	return &value
}

func (t *TextUI) promptBuildTaskRequiredConfig(ctx context.Context, config *build.BuildTaskConfig) error {
	if config == nil {
		return fmt.Errorf("构建任务配置不能为空")
	}

	worldPath, err := t.PromptBuildTaskWorldPath(ctx)
	if err != nil {
		return err
	}
	worldStartPos, err := t.PromptBuildTaskWorldStartPos(ctx)
	if err != nil {
		return err
	}
	worldEndPos, err := t.PromptBuildTaskWorldEndPos(ctx)
	if err != nil {
		return err
	}
	worldDimension, err := t.PromptBuildTaskWorldDimension(ctx, "请选择建筑区域维度：\n")
	if err != nil {
		return err
	}
	startPos, err := t.PromptBuildTaskStartPos(ctx)
	if err != nil {
		return err
	}
	dimension, err := t.PromptBuildTaskDimension(ctx, "请选择构建维度：\n")
	if err != nil {
		return err
	}

	config.WorldPath = worldPath
	config.WorldStartPos = worldStartPos
	config.WorldEndPos = worldEndPos
	config.WorldDimension = worldDimension
	config.StartPos = startPos
	config.Dimension = dimension
	return nil
}
