package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/HonLaderDev/HonLader/cmd/tui/control"
	"github.com/HonLaderDev/HonLader/cmd/tui/ui"
	"github.com/HonLaderDev/HonLader/frame"
	"github.com/HonLaderDev/HonLader/utils/storage"
)

func main() {
	checkTimeBomb()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	textControl := control.NewDefaultTextUIControl(os.Stdin, os.Stdout)

	go func() {
		if err := textControl.Run(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "读取输入失败：%v\n", err)
			os.Exit(1)
		}
	}()

	taskFrame := frame.TaskFrameConfig{Embedded: true}.New(nil)
	launcher := frame.NewLauncher(taskFrame, storage.NewDefaultStorage())
	textUI := ui.NewTextUI(launcher, textControl)
	go func() {
		<-ctx.Done()
		_ = launcher.Stop()
	}()

	if err := textUI.Run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "TUI 运行失败：%v\n", err)
		os.Exit(1)
	}
}
