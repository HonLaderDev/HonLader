package ui

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/HonLaderDev/HonLader/cmd/tui/ui/server_config_tui"
	"github.com/HonLaderDev/HonLader/define"
	build "github.com/HonLaderDev/HonLader/frame/task/build"
)

// StartBuild 选择服务器配置、创建构建任务配置并启动构建任务。
func (t *TextUI) StartBuild(ctx context.Context) error {
	authConfig, err := t.authTUI.EnsureAuthConfig(ctx)
	if err != nil {
		return err
	}

	if checkpoint, ok, err := t.latestBuildCheckpoint(); err != nil {
		return err
	} else if ok {
		confirm, err := t.c.Confirm(ctx, fmt.Sprintf("发现断点 %s，是否恢复[Y/n]：", checkpoint.Metadata.Name), true)
		if err != nil {
			return err
		}
		if confirm {
			return t.startBuildFromCheckpoint(ctx, authConfig, checkpoint)
		}
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
	t.buildProgressTUI.Register(t.l.TaskFrame())
	t.buildProgressTUI.Reset()
	if err := t.l.WatchLog(ctx, NewTextUILogWriter(t)); err != nil {
		return fmt.Errorf("监听 Core 日志失败：%w", err)
	}

	config, err := t.buildTaskTUI.CreateBuildTaskConfig(ctx)
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
	return t.startTaskFrame(ctx)
}

func (t *TextUI) latestBuildCheckpoint() (define.CheckpointConfig, bool, error) {
	checkpoints, err := t.l.DataManager().ListCheckpointConfigs()
	if err != nil {
		return define.CheckpointConfig{}, false, err
	}
	if len(checkpoints) == 0 {
		return define.CheckpointConfig{}, false, nil
	}
	list := make([]define.CheckpointConfig, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		list = append(list, checkpoint)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Metadata.UpdatedAt.After(list[j].Metadata.UpdatedAt)
	})
	for _, checkpoint := range list {
		if len(checkpoint.Tasks) == 0 {
			continue
		}
		if checkpoint.Tasks[0].TaskName == build.Name {
			return checkpoint, true, nil
		}
	}
	return define.CheckpointConfig{}, false, nil
}

func (t *TextUI) startBuildFromCheckpoint(ctx context.Context, authConfig define.Config, checkpoint define.CheckpointConfig) error {
	if len(checkpoint.Tasks) == 0 {
		return fmt.Errorf("断点 %s 没有任务数据", checkpoint.Metadata.Name)
	}
	task, err := build.NewTaskFromInfo(checkpoint.Tasks[0], t.l.TaskFrame())
	if err != nil {
		return err
	}
	t.buildProgressTUI.Register(t.l.TaskFrame())
	t.buildProgressTUI.Reset()
	if err := t.l.WatchLog(ctx, NewTextUILogWriter(t)); err != nil {
		return fmt.Errorf("监听 Core 日志失败：%w", err)
	}
	if err := t.l.DataManager().SaveConfig(authConfig); err != nil {
		return err
	}
	if err := t.l.Connect(ctx, checkpoint.Server); err != nil {
		return fmt.Errorf("连接 Core 失败：%w", err)
	}
	t.l.AddTask(task)
	return t.startTaskFrame(ctx)
}

func (t *TextUI) startTaskFrame(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		done <- t.l.Start()
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		_ = t.l.Stop()
		err := <-done
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return ctx.Err()
	}
}
