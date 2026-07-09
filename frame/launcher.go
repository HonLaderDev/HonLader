package frame

import (
	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils/data"
	"github.com/spf13/afero"
)

// Launcher 聚合任务框架、存储和数据管理器。
type Launcher struct {
	define.TaskFrame
	storage     define.Storage
	dataManager define.DataManager
}

// NewLauncher 基于任务框架和存储创建启动器。
func NewLauncher(taskFrame define.TaskFrame, storage define.Storage) *Launcher {
	launcher := &Launcher{
		TaskFrame:   taskFrame,
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

// Storage 返回底层存储实现。
func (l *Launcher) Storage() define.Storage {
	return l.storage
}

// FileSystem 返回底层文件系统。
func (l *Launcher) FileSystem() afero.Fs {
	return l.storage.FileSystem()
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

var _ define.Launcher = (*Launcher)(nil)
