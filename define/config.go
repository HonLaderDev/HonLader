package define

import "time"

// Config 保存 HonLader 全局配置。
type Config struct {
	// AuthServer 是认证服务地址。
	AuthServer string `mapstructure:"auth_server"`
	// AuthToken 是认证令牌或登录结果令牌。
	AuthToken string `mapstructure:"auth_token"`
	// AuthMode 是认证模式名称。
	AuthMode string `mapstructure:"auth_mode"`
	// AuthCookie 是需要持久化的认证 Cookie。
	AuthCookie string `mapstructure:"auth_cookie"`
}

// Metadata 保存本地配置对象的通用元信息。
type Metadata struct {
	// Name 是配置名称，也是本地配置的查找键。
	Name string `mapstructure:"name"`
	// CreatedAt 是配置首次创建时间。
	CreatedAt time.Time `mapstructure:"created_at"`
	// UpdatedAt 是配置最近更新时间。
	UpdatedAt time.Time `mapstructure:"updated_at"`
}

// ServerConfig 保存连接服务器需要的基础配置。
type ServerConfig struct {
	// Metadata 是服务器配置的本地元信息。
	Metadata Metadata `mapstructure:"metadata"`
	// ServerCode 是目标租赁服或服务器连接码。
	ServerCode string `mapstructure:"server_code"`
	// ServerPassword 是目标服务器连接密码。
	ServerPassword string `mapstructure:"server_password"`
}

// ConnectConfig 保存连接 Core 所需的完整配置。
type ConnectConfig struct {
	// AuthServer 是认证服务地址。
	AuthServer string `mapstructure:"auth_server"`
	// AuthToken 是认证令牌或登录结果令牌。
	AuthToken string `mapstructure:"auth_token"`
	// ServerCode 是目标租赁服或服务器连接码。
	ServerCode string `mapstructure:"server_code"`
	// ServerPassword 是目标服务器连接密码。
	ServerPassword string `mapstructure:"server_password"`
}

// TaskConfig 保存单个任务配置。
type TaskConfig = map[string]any

// TaskCheckpoint 保存单个任务断点。
type TaskCheckpoint = map[string]any

// TaskInfo 保存单个任务的可持久化信息。
type TaskInfo struct {
	// TaskName 是任务类型名称。
	TaskName string `mapstructure:"task_name"`
	// Config 是任务的配置数据。
	Config TaskConfig `mapstructure:"config"`
	// Checkpoint 是任务的运行断点数据。
	Checkpoint TaskCheckpoint `mapstructure:"checkpoint"`
}

// TaskGroupCheckpoint 保存一组任务配置和对应断点。
type TaskGroupCheckpoint struct {
	// TaskConfigs 是任务组内的任务配置列表。
	TaskConfigs []TaskConfig `mapstructure:"task_configs"`
	// TaskCheckpoints 是与 TaskConfigs 顺序对应的任务断点列表。
	TaskCheckpoints []TaskCheckpoint `mapstructure:"task_checkpoints"`
	// Server 是断点关联的服务器配置。
	Server ServerConfig `mapstructure:"server"`
	// CurrentTaskIndex 是当前执行到的任务索引。
	CurrentTaskIndex int `mapstructure:"current_task_index"`
}

// TaskGroupConfig 保存一组任务配置。
type TaskGroupConfig struct {
	// Metadata 是任务组配置的本地元信息。
	Metadata Metadata `mapstructure:"metadata"`
	// Tasks 是任务组内的任务配置列表。
	Tasks []TaskConfig `mapstructure:"tasks"`
}

// CheckpointConfig 保存任务组运行断点。
type CheckpointConfig struct {
	// Metadata 是断点配置的本地元信息。
	Metadata Metadata `mapstructure:"metadata"`
	// Server 是断点关联的服务器配置。
	Server ServerConfig `mapstructure:"server"`
	// Tasks 是任务组内的可持久化任务信息列表。
	Tasks []TaskInfo `mapstructure:"tasks"`
	// CurrentTaskIndex 是当前执行到的任务索引。
	CurrentTaskIndex int `mapstructure:"current_task_index"`
}
