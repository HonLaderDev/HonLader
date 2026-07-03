package define

import "time"

// ServerConfig 保存连接服务器需要的基础配置。
type ServerConfig struct {
	Name           string    `mapstructure:"name"`
	CreatedAt      time.Time `mapstructure:"created_at"`
	UpdatedAt      time.Time `mapstructure:"updated_at"`
	ServerCode     string    `mapstructure:"server_code"`
	ServerPassword string    `mapstructure:"server_password"`
}

// TaskConfig 保存单个任务配置。
type TaskConfig = map[string]any

// TaskCheckpoint 保存单个任务断点。
type TaskCheckpoint = map[string]any

// TaskInfo 保存单个任务的可持久化信息。
type TaskInfo struct {
	Config     TaskConfig     `mapstructure:"config"`
	Checkpoint TaskCheckpoint `mapstructure:"checkpoint"`
}

// TaskGroupConfig 保存一组任务配置。
type TaskGroupConfig struct {
	Name      string       `mapstructure:"name"`
	CreatedAt time.Time    `mapstructure:"created_at"`
	UpdatedAt time.Time    `mapstructure:"updated_at"`
	Tasks     []TaskConfig `mapstructure:"tasks"`
}

// CheckpointConfig 保存任务组运行断点。
type CheckpointConfig struct {
	Name             string       `mapstructure:"name"`
	CreatedAt        time.Time    `mapstructure:"created_at"`
	UpdatedAt        time.Time    `mapstructure:"updated_at"`
	Server           ServerConfig `mapstructure:"server"`
	Tasks            []TaskInfo   `mapstructure:"tasks"`
	CurrentTaskIndex int          `mapstructure:"current_task_index"`
}
