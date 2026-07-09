package main

import (
	"context"
	"fmt"
	"os"

	"github.com/HonLaderDev/HonLader/cmd/tui/control"
)

func main() {
	control := control.NewDefaultTextUIControl(os.Stdin, os.Stdout)

	// 启动阻塞主循环（后台持续读取输入）
	go func() {
		if err := control.Run(); err != nil {
			panic(err)
		}
	}()

	ctx := context.Background()
	control.Print("=== TUI Demo ===\n")

	name, _ := control.Prompt(ctx, "请输入姓名：")
	control.Print(fmt.Sprintf("你好：%s\n", name))

	sel, _ := control.Select(ctx, "请选择操作：\n", []string{"新增", "查询", "删除"})
	control.Print(fmt.Sprintf("选中索引: %d\n", sel))

	ok, _ := control.Confirm(ctx, "确认保存？[Y/n]", true)
	control.Print(fmt.Sprintf("确认：%v\n", ok))
}
