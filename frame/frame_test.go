package frame

import (
	"testing"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils/storage"
)

type noopTask struct {
	frame define.TaskFrame
}

func (t noopTask) Name() string                { return "noop" }
func (t noopTask) TaskFrame() define.TaskFrame { return t.frame }
func (t noopTask) Start() error                { return nil }
func (t noopTask) Pause() error                { return nil }
func (t noopTask) Resume() error               { return nil }
func (t noopTask) Stop() error                 { return nil }

func TestTaskFrameStartGeneratesTaskGroupName(t *testing.T) {
	taskFrame := TaskFrameConfig{}.New(nil)
	taskFrame.AddTask(noopTask{frame: taskFrame})

	if err := taskFrame.Start(); err != nil {
		t.Fatalf("start task frame: %v", err)
	}
	if taskFrame.TaskGroupName() == "" {
		t.Fatal("task group name is empty")
	}
}

func TestLauncherCheckpointConfigUsesCurrentServerConfig(t *testing.T) {
	taskFrame := TaskFrameConfig{}.New(nil)
	taskFrame.serverConfig = define.ServerConfig{
		ServerCode:     "server-code",
		ServerPassword: "server-password",
	}
	launcher := NewLauncher(taskFrame, storage.NewDefaultStorage())

	config, err := launcher.checkpointConfig()
	if err != nil {
		t.Fatalf("checkpoint config: %v", err)
	}
	if config.Server.ServerCode != "server-code" {
		t.Fatalf("server code = %q, want server-code", config.Server.ServerCode)
	}
	if config.Server.ServerPassword != "server-password" {
		t.Fatalf("server password = %q, want server-password", config.Server.ServerPassword)
	}
}
