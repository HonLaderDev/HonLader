package define

import "github.com/spf13/afero"

// DataManager 管理本地任务、服务器和断点配置。
type DataManager interface {
	FileSystem() afero.Fs
	Storage() Storage

	ConfigDir() string
	DataDir() string
	TmpDir() string
	TasksDir() string
	CheckpointsDir() string
	ServersDir() string
	BuildingsDir() string
	ExportsDir() string

	EnsureLayout() error

	SaveServerConfig(config ServerConfig) error
	LoadServerConfig(name string) (ServerConfig, bool, error)
	ListServerConfigs() (map[string]ServerConfig, error)
	DeleteServerConfig(name string) (bool, error)

	SaveTaskConfig(name string, config TaskGroupConfig) error
	LoadTaskConfig(name string) (TaskGroupConfig, bool, error)
	ListTaskConfigs() (map[string]TaskGroupConfig, error)

	SaveCheckpointConfig(name string, config CheckpointConfig) error
	LoadCheckpointConfig(name string) (CheckpointConfig, bool, error)
	ListCheckpointConfigs() (map[string]CheckpointConfig, error)
}
