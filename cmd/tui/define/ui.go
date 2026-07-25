package define

import (
	"context"

	core_define "github.com/HonLaderDev/HonLader/define"
)

// TextUI 表示可运行的文本交互界面。
type TextUI interface {
	TextUIUtils

	// Run 运行文本交互界面。
	Run(ctx context.Context) error

	// Control 返回文本交互控制器。
	Control() TextUIControl

	// DataManager 返回本地配置数据管理器。
	DataManager() core_define.DataManager
}

// TextUIUtils 提供 TUI 通用输入和格式化能力。
type TextUIUtils interface {
	// Format 统一格式化界面展示值。
	Format(value any) string

	// PromptRequired 读取必填文本，空输入时使用可选默认值。
	PromptRequired(ctx context.Context, hint string, defaultValues ...string) (string, error)

	// PromptBlockPos 读取并解析 x,y,z 格式的方块坐标。
	PromptBlockPos(ctx context.Context, hint string, defaultValues ...core_define.BlockPos) (core_define.BlockPos, error)

	// PromptDimension 读取并解析整数维度 ID。
	PromptDimension(ctx context.Context, hint string, defaultValues ...core_define.Dimension) (core_define.Dimension, error)

	// PromptInt 读取并解析整数输入。
	PromptInt(ctx context.Context, hint string, defaultValues ...int) (int, error)

	// PromptFloat 读取并解析浮点数输入。
	PromptFloat(ctx context.Context, hint string, defaultValues ...float64) (float64, error)
}

// TextUIControl 通用终端文本交互界面接口。
type TextUIControl interface {
	// Print 打印消息。
	Print(msg string)

	// Prompt 交互式问答输入，返回用户输入。
	Prompt(ctx context.Context, hint string) (string, error)

	// Select 数字菜单选择。
	Select(ctx context.Context, title string, options []string) (int, error)

	// Confirm 确认弹窗 Y/n。
	Confirm(ctx context.Context, hint string, def bool) (bool, error)

	// Run 阻塞运行主 IO 循环，接管输入读取，统一处理 ctx 取消。
	Run() error
}
