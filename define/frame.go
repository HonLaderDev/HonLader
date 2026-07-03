package define

import (
	"context"

	frame_api "github.com/RedLaderDev/RedLader-core-api/frame"
	"github.com/asaskevich/EventBus"
)

type TaskFrame interface {
	Client() frame_api.Client
	EventBus() EventBus.Bus
	// 任务组名称
	TaskGroupName() string
	// 设置任务组名称
	WithTaskGroupName(name string) TaskFrame
	// 所有任务
	Tasks() []Task
	// 当前任务索引
	CurrentTaskIndex() int
	// 连接并检查 Core 可用性
	Connect(ctx context.Context) error
	// 添加任务
	AddTask(task Task) TaskFrame
	// 运行所有任务
	Start() error
	// 暂停当前任务
	Pause() error
	// 恢复当前任务并继续剩余任务
	Resume() error
	// 停止当前任务
	Stop() error
	// 停止并关闭 Client
	Close() error
}

type Task interface {
	Name() string
	TaskFrame() TaskFrame
	Start() error
	Pause() error
	Resume() error
	Stop() error
}
