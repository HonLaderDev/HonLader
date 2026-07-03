package data

import (
	"path/filepath"

	"github.com/RedLaderDev/Fatalder/consts"
	"github.com/RedLaderDev/Fatalder/define"
	"github.com/RedLaderDev/Fatalder/utils/storage"
	"github.com/spf13/afero"
)

// DataManager 管理本地数据目录。
type DataManager struct {
	storage define.Storage

	configDir string
	dataDir   string
	tmpDir    string

	tasksDir       string
	checkpointsDir string
	serversDir     string
	buildingsDir   string
	exportsDir     string
}

// NewDataManager 基于 Storage 创建数据管理器。
func NewDataManager(s define.Storage) *DataManager {
	if s == nil {
		s = storage.NewDefaultStorage()
	}

	configDir := filepath.Join(s.ConfigDir(), consts.Name)
	dataDir := filepath.Join(s.DataDir(), consts.Name)
	manager := &DataManager{
		storage: s,

		configDir: configDir,
		dataDir:   dataDir,
		tmpDir:    filepath.Join(dataDir, TmpDir),

		tasksDir:       filepath.Join(configDir, TasksDir),
		checkpointsDir: filepath.Join(configDir, CheckpointsDir),
		serversDir:     filepath.Join(configDir, ServersDir),
		buildingsDir:   filepath.Join(configDir, BuildingsDir),
		exportsDir:     filepath.Join(dataDir, ExportsDir),
	}
	_ = manager.EnsureLayout()
	return manager
}

// FileSystem 返回 DataManager 使用的文件系统。
func (m *DataManager) FileSystem() afero.Fs {
	return m.storage.FileSystem()
}

// Storage 返回 DataManager 使用的存储后端。
func (m *DataManager) Storage() define.Storage {
	return m.storage
}

// ConfigDir 返回配置目录。
func (m *DataManager) ConfigDir() string {
	return m.configDir
}

// DataDir 返回数据目录。
func (m *DataManager) DataDir() string {
	return m.dataDir
}

// TmpDir 返回临时目录。
func (m *DataManager) TmpDir() string {
	return m.tmpDir
}

// TasksDir 返回任务配置目录。
func (m *DataManager) TasksDir() string {
	return m.tasksDir
}

// CheckpointsDir 返回断点目录。
func (m *DataManager) CheckpointsDir() string {
	return m.checkpointsDir
}

// ServersDir 返回服务器配置目录。
func (m *DataManager) ServersDir() string {
	return m.serversDir
}

// BuildingsDir 返回建筑文件目录。
func (m *DataManager) BuildingsDir() string {
	return m.buildingsDir
}

// ExportsDir 返回导出文件目录。
func (m *DataManager) ExportsDir() string {
	return m.exportsDir
}

// EnsureLayout 确保基础数据目录存在。
func (m *DataManager) EnsureLayout() error {
	fs := m.FileSystem()
	for _, dir := range []string{
		m.configDir,
		m.dataDir,
		m.tmpDir,
		m.tasksDir,
		m.checkpointsDir,
		m.serversDir,
		m.buildingsDir,
		m.exportsDir,
	} {
		if err := fs.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
