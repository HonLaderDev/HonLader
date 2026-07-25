package export_task_tui

import (
	"context"
	"fmt"

	tui_define "github.com/HonLaderDev/HonLader/cmd/tui/define"
	export "github.com/HonLaderDev/HonLader/frame/task/export"
)

// ExportTaskTUI 负责导出任务配置相关的文本交互。
type ExportTaskTUI struct {
	tui tui_define.TextUI
}

// NewExportTaskTUI 创建导出任务配置交互实例。
func NewExportTaskTUI(tui tui_define.TextUI) *ExportTaskTUI {
	return &ExportTaskTUI{tui: tui}
}

// CreateExportTaskConfig 创建导出任务配置。
func (t *ExportTaskTUI) CreateExportTaskConfig(ctx context.Context) (export.ExportTaskConfig, error) {
	config := export.ExportTaskConfig{}
	config.FillDefault()

	filePath, err := t.tui.PromptRequired(ctx, "请输入导出文件路径：")
	if err != nil {
		return export.ExportTaskConfig{}, err
	}
	startPos, err := t.tui.PromptBlockPos(ctx, "请输入导出区域起点(x,y,z)：")
	if err != nil {
		return export.ExportTaskConfig{}, err
	}
	endPos, err := t.tui.PromptBlockPos(ctx, "请输入导出区域终点(x,y,z)：")
	if err != nil {
		return export.ExportTaskConfig{}, err
	}
	dimension, err := t.tui.PromptDimension(ctx, "请输入导出维度：")
	if err != nil {
		return export.ExportTaskConfig{}, err
	}
	batchSide, err := t.tui.PromptInt(ctx, "请输入批次区块边长(默认 16)：", config.ChunkBatchSide)
	if err != nil {
		return export.ExportTaskConfig{}, err
	}

	config.FilePath = filePath
	config.StartPos = startPos
	config.EndPos = endPos
	config.Dimension = dimension
	config.ChunkBatchSide = batchSide

	t.PrintExportTaskSummary(config)
	confirm, err := t.tui.Control().Confirm(ctx, "是否进行导出[Y/n]：", true)
	if err != nil {
		return export.ExportTaskConfig{}, err
	}
	if confirm {
		return config, nil
	}
	return export.ExportTaskConfig{}, fmt.Errorf("已取消导出")
}

// PrintExportTaskSummary 打印导出任务配置摘要。
func (t *ExportTaskTUI) PrintExportTaskSummary(config export.ExportTaskConfig) {
	t.tui.Control().Print("导出任务配置摘要：\n")
	t.tui.Control().Print(fmt.Sprintf("  导出文件：%s\n", config.FilePath))
	t.tui.Control().Print(fmt.Sprintf("  导出起点：%s\n", t.tui.Format(config.StartPos)))
	t.tui.Control().Print(fmt.Sprintf("  导出终点：%s\n", t.tui.Format(config.EndPos)))
	t.tui.Control().Print(fmt.Sprintf("  导出维度：%s\n", t.tui.Format(config.Dimension)))
	t.tui.Control().Print(fmt.Sprintf("  批次边长：%d\n", config.ChunkBatchSide))
}
