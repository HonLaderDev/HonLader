package control

import "context"

// TextUIControl 通用终端文本交互界面接口
type TextUIControl interface {
	// Print 打印消息
	Print(msg string)

	// Prompt 交互式问答输入，返回用户输入
	Prompt(ctx context.Context, hint string) (string, error)

	// Select 数字菜单选择
	Select(ctx context.Context, title string, options []string) (int, error)

	// Confirm 确认弹窗 Y/n
	Confirm(ctx context.Context, hint string, def bool) (bool, error)

	// Run 阻塞运行主IO循环，接管输入读取，统一处理ctx取消
	Run() error
}
