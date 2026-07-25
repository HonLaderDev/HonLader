package define

import (
	"context"
	"io"

	"github.com/asaskevich/EventBus"
)

// Launcher 聚合任务运行框架、存储和数据管理能力。
type Launcher interface {
	Storage
	TaskFrame() TaskFrame
	EventBus() EventBus.Bus
	TaskGroupName() string
	Tasks() []Task
	CurrentTaskIndex() int
	Connect(ctx context.Context, server ServerConfig) error
	WatchLog(ctx context.Context, writer io.Writer) error
	AddTask(task Task) Launcher
	Start() error
	Pause() error
	Resume() error
	Stop() error
	Close() error
	DataManager() DataManager
	LoadTaskGroup(tasks []Task) error
	LoadTaskGroupCheckpoint(checkpoint TaskGroupCheckpoint) error
}
