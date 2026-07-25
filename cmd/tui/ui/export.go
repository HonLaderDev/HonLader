package ui

import (
	"context"
	"fmt"

	"github.com/HonLaderDev/HonLader/cmd/tui/ui/server_config_tui"
)

// StartExport 选择服务器配置、创建导出任务配置并启动导出任务。
func (t *TextUI) StartExport(ctx context.Context) error {
	authConfig, err := t.authTUI.EnsureAuthConfig(ctx)
	if err != nil {
		return err
	}

	serverConfig, result, err := t.serverConfigTUI.SelectServerConfigForTask(ctx)
	if err != nil {
		return err
	}
	if result == server_config_tui.ServerConfigSelectReturn {
		return nil
	}
	if err := t.l.DataManager().SaveLatestServerConfig(serverConfig); err != nil {
		return err
	}

	config, err := t.exportTaskTUI.CreateExportTaskConfig(ctx)
	if err != nil {
		return err
	}
	if err := t.l.DataManager().SaveConfig(authConfig); err != nil {
		return err
	}
	if err := t.l.Connect(ctx, serverConfig); err != nil {
		return fmt.Errorf("连接 Core 失败：%w", err)
	}

	t.l.AddTask(config.NewTask(t.l.TaskFrame()))
	return t.l.Start()
}
