package ui

import (
	"context"
	"fmt"

	tui_define "github.com/HonLaderDev/HonLader/cmd/tui/define"
	"github.com/HonLaderDev/HonLader/cmd/tui/ui/auth_tui"
	"github.com/HonLaderDev/HonLader/cmd/tui/ui/build_progress_tui"
	"github.com/HonLaderDev/HonLader/cmd/tui/ui/build_task_tui"
	"github.com/HonLaderDev/HonLader/cmd/tui/ui/config_tui"
	"github.com/HonLaderDev/HonLader/cmd/tui/ui/export_task_tui"
	"github.com/HonLaderDev/HonLader/cmd/tui/ui/server_config_tui"
	core_define "github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/frame"
)

type TextUI struct {
	l                core_define.Launcher
	c                tui_define.TextUIControl
	buildProgressTUI *build_progress_tui.BuildProgressTUI
	buildTaskTUI     *build_task_tui.BuildTaskTUI
	exportTaskTUI    *export_task_tui.ExportTaskTUI
	configTUI        *config_tui.ConfigTUI
	serverConfigTUI  *server_config_tui.ServerConfigTUI
	authTUI          *auth_tui.AuthTUI
}

var _ tui_define.TextUI = (*TextUI)(nil)

// NewTextUI 创建命令行交互界面实例。
func NewTextUI(launcher core_define.Launcher, control tui_define.TextUIControl) *TextUI {
	tui := &TextUI{
		l: launcher,
		c: control,
	}
	tui.buildProgressTUI = build_progress_tui.NewBuildProgressTUI(tui)
	tui.buildTaskTUI = build_task_tui.NewBuildTaskTUI(tui)
	tui.exportTaskTUI = export_task_tui.NewExportTaskTUI(tui)
	tui.configTUI = config_tui.NewConfigTUI(tui)
	tui.serverConfigTUI = server_config_tui.NewServerConfigTUI(tui)
	tui.authTUI = auth_tui.NewAuthTUI(tui)
	_ = launcher.EventBus().Subscribe(frame.EventNameLauncherCheckpointFailed, func(err error) {
		tui.Control().Print(fmt.Sprintf("断点保存失败：%v\n", err))
	})
	return tui
}

// Control 返回文本交互控制器。
func (t *TextUI) Control() tui_define.TextUIControl {
	return t.c
}

// DataManager 返回本地配置数据管理器。
func (t *TextUI) DataManager() core_define.DataManager {
	return t.l.DataManager()
}

// Run 显示 TUI 主菜单并分发到对应任务流程。
func (t *TextUI) Run(ctx context.Context) error {
	for {
		sel, err := t.c.Select(
			ctx,
			"请选择操作：\n",
			[]string{
				"开始构建",
				"开始导出",
				"服务器配置",
				"机器人配置",
				"退出程序",
			},
		)
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			if err := t.StartBuild(ctx); err != nil {
				return err
			}
		case 1:
			if err := t.StartExport(ctx); err != nil {
				return err
			}
		case 2:
			if err := t.serverConfigTUI.Run(ctx); err != nil {
				return err
			}
		case 3:
			if err := t.authTUI.Run(ctx); err != nil {
				return err
			}
		case 4:
			return nil
		default:
			return fmt.Errorf("未知操作：%d", sel)
		}
	}
}
