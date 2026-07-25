package frame

import (
	"context"
	"fmt"
	"io"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils/data"
	"github.com/asaskevich/EventBus"
	"github.com/spf13/afero"
)

// Launcher 聚合任务框架、存储和数据管理器。
type Launcher struct {
	taskFrame   define.TaskFrame
	storage     define.Storage
	dataManager define.DataManager
}

// NewLauncher 基于任务框架和存储创建启动器。
func NewLauncher(taskFrame define.TaskFrame, storage define.Storage) *Launcher {
	launcher := &Launcher{
		taskFrame:   taskFrame,
		storage:     storage,
		dataManager: data.NewDataManager(storage),
	}
	launcher.watchTaskCheckpoints()
	return launcher
}

// DataManager 返回数据管理器。
func (l *Launcher) DataManager() define.DataManager {
	return l.dataManager
}

// TaskFrame 返回底层任务运行框架。
func (l *Launcher) TaskFrame() define.TaskFrame {
	return l.taskFrame
}

// EventBus 返回框架事件总线。
func (l *Launcher) EventBus() EventBus.Bus {
	return l.taskFrame.EventBus()
}

// TaskGroupName 返回当前任务组名称。
func (l *Launcher) TaskGroupName() string {
	return l.taskFrame.TaskGroupName()
}

// Tasks 返回当前框架持有的任务列表。
func (l *Launcher) Tasks() []define.Task {
	return l.taskFrame.Tasks()
}

// CurrentTaskIndex 返回当前正在处理的任务索引。
func (l *Launcher) CurrentTaskIndex() int {
	return l.taskFrame.CurrentTaskIndex()
}

// Storage 返回底层存储实现。
func (l *Launcher) Storage() define.Storage {
	return l.storage
}

// FileSystem 返回底层文件系统。
func (l *Launcher) FileSystem() afero.Fs {
	return l.storage.FileSystem()
}

// LoadTaskGroup 将任务列表加载到 Launcher。
func (l *Launcher) LoadTaskGroup(tasks []define.Task) error {
	for _, task := range tasks {
		l.AddTask(task)
	}
	return nil
}

// LoadTaskGroupCheckpoint 将任务组断点加载到 Launcher。
func (l *Launcher) LoadTaskGroupCheckpoint(checkpoint define.TaskGroupCheckpoint) error {
	return fmt.Errorf("Launcher.LoadTaskGroupCheckpoint: task factory is required, task_configs=%d", len(checkpoint.TaskConfigs))
}

// Connect 使用当前保存的认证配置连接指定服务器。
func (l *Launcher) Connect(ctx context.Context, server define.ServerConfig) error {
	config, ok, err := l.dataManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("Launcher.Connect: load auth config: %w", err)
	}
	if !ok {
		return fmt.Errorf("Launcher.Connect: auth config not found")
	}
	return l.taskFrame.Connect(ctx, define.ConnectConfig{
		AuthServer:     config.AuthServer,
		AuthToken:      config.AuthToken,
		ServerCode:     server.ServerCode,
		ServerPassword: server.ServerPassword,
	})
}

// AddTask 添加任务并返回 Launcher 自身。
func (l *Launcher) AddTask(task define.Task) define.Launcher {
	l.taskFrame.AddTask(task)
	return l
}

// WatchLog 监听 Core 日志并写入指定 Writer。
func (l *Launcher) WatchLog(ctx context.Context, writer io.Writer) error {
	return l.taskFrame.WatchLog(ctx, writer)
}

// Start 按添加顺序启动所有任务。
func (l *Launcher) Start() error {
	return l.taskFrame.Start()
}

// Pause 暂停当前任务。
func (l *Launcher) Pause() error {
	return l.taskFrame.Pause()
}

// Resume 恢复当前任务并继续剩余任务。
func (l *Launcher) Resume() error {
	return l.taskFrame.Resume()
}

// ConfigDir 返回应用配置目录。
func (l *Launcher) ConfigDir() string {
	return l.dataManager.ConfigDir()
}

// DataDir 返回应用数据目录。
func (l *Launcher) DataDir() string {
	return l.dataManager.DataDir()
}

// TmpDir 返回应用临时目录。
func (l *Launcher) TmpDir() string {
	return l.dataManager.TmpDir()
}

// DownloadDir 返回用户下载目录。
func (l *Launcher) DownloadDir() string {
	return l.storage.DownloadDir()
}

// Stop 停止当前任务，并同步保存一次断点。
func (l *Launcher) Stop() error {
	if err := l.taskFrame.Stop(); err != nil {
		return fmt.Errorf("Launcher.Stop: %w", err)
	}
	l.saveCheckpoint()
	return nil
}

// Close 停止所有任务、同步保存断点并关闭 Core 连接。
func (l *Launcher) Close() error {
	if err := l.taskFrame.Close(); err != nil {
		return fmt.Errorf("Launcher.Close: %w", err)
	}
	l.saveCheckpoint()
	return nil
}

var _ define.Launcher = (*Launcher)(nil)
