package frame

import (
	"context"
	"fmt"

	frame_api "github.com/RedLaderDev/RedLader-core-api/frame"
	client "github.com/RedLaderDev/RedLader-core-client"
	"github.com/RedLaderDev/RedLader/define"
	"github.com/asaskevich/EventBus"
)

// TaskFrame 是任务运行框架实现，负责持有客户端、事件总线和任务列表。
type TaskFrame struct {
	client           frame_api.Client
	closer           func() error
	eventBus         EventBus.Bus
	taskGroupName    string
	tasks            []define.Task
	currentTaskIndex int
	config           TaskFrameConfig
}

type ClientConfig = client.FrameConfig

// TaskFrameConfig 描述 TaskFrame 的创建参数。
type TaskFrameConfig struct {
	ClientConfig
	// Embedded 标记是否使用嵌入式运行模式，当前暂不参与逻辑。
	Embedded bool
}

// New 创建一个默认事件总线的 TaskFrame。
func (c TaskFrameConfig) New(coreClient frame_api.Client) *TaskFrame {
	return &TaskFrame{
		client:   coreClient,
		eventBus: EventBus.New(),
		config:   c,
	}
}

// Client 返回底层 RedLader Core 客户端。
func (f *TaskFrame) Client() frame_api.Client {
	return f.client
}

// EventBus 返回框架事件总线。
func (f *TaskFrame) EventBus() EventBus.Bus {
	return f.eventBus
}

// TaskGroupName 返回当前任务组名称。
func (f *TaskFrame) TaskGroupName() string {
	return f.taskGroupName
}

// WithTaskGroupName 设置任务组名称并返回自身，便于链式调用。
func (f *TaskFrame) WithTaskGroupName(name string) define.TaskFrame {
	f.taskGroupName = name
	return f
}

// Tasks 返回当前框架持有的任务列表。
func (f *TaskFrame) Tasks() []define.Task {
	return f.tasks
}

// CurrentTaskIndex 返回当前正在处理的任务索引。
func (f *TaskFrame) CurrentTaskIndex() int {
	return f.currentTaskIndex
}

// Connect 确保 Core 客户端已经连接。
func (f *TaskFrame) Connect(ctx context.Context) error {
	if err := f.initClient(); err != nil {
		return fmt.Errorf("TaskFrame.Connect: init client: %w", err)
	}
	state, err := f.client.Frame().GetConnectionState(ctx)
	if err != nil {
		return fmt.Errorf("TaskFrame.Connect: get connection state: %w", err)
	}
	if state.Connected {
		return nil
	}

	if _, err := f.client.Frame().StartConnection(ctx, f.config.ClientConfig); err != nil {
		return fmt.Errorf("TaskFrame.Connect: start connection: %w", err)
	}

	state, err = f.client.Frame().GetConnectionState(ctx)
	if err != nil {
		return fmt.Errorf("TaskFrame.Connect: get connection state after start: %w", err)
	}
	if !state.Connected {
		return fmt.Errorf("TaskFrame.Connect: core disconnected after start: %s", state.CloseReason)
	}
	return nil
}

// AddTask 添加任务到框架并返回自身，便于链式调用。
func (f *TaskFrame) AddTask(task define.Task) define.TaskFrame {
	f.tasks = append(f.tasks, task)
	f.publish(EventNameTaskFrameTaskAdded, len(f.tasks)-1)
	return f
}

// Start 按添加顺序启动所有任务。
func (f *TaskFrame) Start() error {
	f.publish(EventNameTaskFrameStart, len(f.tasks))
	for i, task := range f.tasks {
		if i > 0 {
			f.publish(EventNameTaskFrameNextTask, i)
		}
		f.currentTaskIndex = i
		f.publish(EventNameTaskFrameTaskStart, i)
		if err := task.Start(); err != nil {
			f.publish(EventNameTaskFrameTaskFailed, i, err)
			return fmt.Errorf("TaskFrame.Start: start task %q: %w", task.Name(), err)
		}
		f.publish(EventNameTaskFrameTaskFinish, i)
	}
	f.publish(EventNameTaskFrameFinish)
	return nil
}

// Pause 暂停当前任务。
func (f *TaskFrame) Pause() error {
	task := f.currentTask()
	if task == nil {
		return nil
	}
	f.publish(EventNameTaskFramePause, f.currentTaskIndex)
	if err := task.Pause(); err != nil {
		return fmt.Errorf("TaskFrame.Pause: pause task %q: %w", task.Name(), err)
	}
	return nil
}

// Resume 恢复当前任务，并在其完成后继续执行剩余任务。
func (f *TaskFrame) Resume() error {
	task := f.currentTask()
	if task == nil {
		return nil
	}
	f.publish(EventNameTaskFrameResume, f.currentTaskIndex)
	if err := task.Resume(); err != nil {
		f.publish(EventNameTaskFrameTaskFailed, f.currentTaskIndex, err)
		return fmt.Errorf("TaskFrame.Resume: resume task %q: %w", task.Name(), err)
	}
	f.publish(EventNameTaskFrameTaskFinish, f.currentTaskIndex)
	for i := f.currentTaskIndex + 1; i < len(f.tasks); i++ {
		f.currentTaskIndex = i
		task = f.tasks[i]
		f.publish(EventNameTaskFrameNextTask, i)
		f.publish(EventNameTaskFrameTaskStart, i)
		if err := task.Start(); err != nil {
			f.publish(EventNameTaskFrameTaskFailed, i, err)
			return fmt.Errorf("TaskFrame.Resume: start task %q: %w", task.Name(), err)
		}
		f.publish(EventNameTaskFrameTaskFinish, i)
	}
	f.publish(EventNameTaskFrameFinish)
	return nil
}

// Stop 停止当前任务，并将当前任务索引重置为 0。
func (f *TaskFrame) Stop() error {
	task := f.currentTask()
	if task == nil {
		f.currentTaskIndex = 0
		return nil
	}
	f.publish(EventNameTaskFrameStop, f.currentTaskIndex)
	if err := task.Stop(); err != nil {
		return fmt.Errorf("TaskFrame.Stop: stop task %q: %w", task.Name(), err)
	}
	f.currentTaskIndex = 0
	return nil
}

// Close 停止所有任务并关闭 Core 连接。
func (f *TaskFrame) Close() error {
	f.publish(EventNameTaskFrameClose)
	if err := f.Stop(); err != nil {
		return fmt.Errorf("TaskFrame.Close: %w", err)
	}
	if f.client != nil {
		if err := f.client.Frame().StopConnection(context.Background()); err != nil {
			return fmt.Errorf("TaskFrame.Close: stop core connection: %w", err)
		}
	}
	if f.closer != nil {
		if err := f.closer(); err != nil {
			return fmt.Errorf("TaskFrame.Close: close core client: %w", err)
		}
		f.closer = nil
	}
	f.client = nil
	return nil
}

func (f *TaskFrame) currentTask() define.Task {
	if f.currentTaskIndex < 0 || f.currentTaskIndex >= len(f.tasks) {
		return nil
	}
	return f.tasks[f.currentTaskIndex]
}

var _ define.TaskFrame = (*TaskFrame)(nil)
