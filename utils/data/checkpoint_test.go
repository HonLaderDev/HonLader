package data

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/spf13/afero"
)

type memoryStorage struct {
	fs afero.Fs
}

func (s memoryStorage) FileSystem() afero.Fs { return s.fs }
func (s memoryStorage) ConfigDir() string    { return "/config" }
func (s memoryStorage) DataDir() string      { return "/data" }
func (s memoryStorage) TmpDir() string       { return "/tmp" }
func (s memoryStorage) DownloadDir() string  { return "/download" }

func TestSaveCheckpointConfigUpdatesExistingCheckpoint(t *testing.T) {
	manager := NewDataManager(memoryStorage{fs: afero.NewMemMapFs()})

	first := define.CheckpointConfig{
		Tasks: []define.TaskInfo{{
			TaskName:   "Build",
			Checkpoint: map[string]any{"current_chunk": 4},
		}},
	}
	if err := manager.SaveCheckpointConfig("build", first); err != nil {
		t.Fatalf("save first checkpoint: %v", err)
	}

	second := define.CheckpointConfig{
		Tasks: []define.TaskInfo{{
			TaskName:   "Build",
			Checkpoint: map[string]any{"current_chunk": 8},
		}},
	}
	if err := manager.SaveCheckpointConfig("build", second); err != nil {
		t.Fatalf("save second checkpoint: %v", err)
	}

	loaded, exists, err := manager.LoadCheckpointConfig("build")
	if err != nil {
		t.Fatalf("load checkpoint: %v", err)
	}
	if !exists {
		t.Fatal("checkpoint does not exist")
	}
	got := loaded.Tasks[0].Checkpoint["current_chunk"]
	if reflect.ValueOf(got).Convert(reflect.TypeOf(int64(0))).Int() != 8 {
		t.Fatalf("current_chunk = %v, want 8", got)
	}

	names, err := listJSONNames(manager.FileSystem(), manager.CheckpointsDir())
	if err != nil {
		t.Fatalf("list checkpoint files: %v", err)
	}
	if len(names) != 1 {
		t.Fatalf("checkpoint file count = %d, want 1", len(names))
	}
	if names[0] != checkpointFileName("build") {
		t.Fatalf("checkpoint file name = %q, want %q", names[0], checkpointFileName("build"))
	}

	data, err := afero.ReadFile(manager.FileSystem(), filepath.Join(manager.CheckpointsDir(), names[0]+jsonFileSuffix))
	if err != nil {
		t.Fatalf("read checkpoint file: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("decode checkpoint json: %v", err)
	}
	tasks := raw["tasks"].([]any)
	task := tasks[0].(map[string]any)
	if _, ok := task["task_name"]; !ok {
		t.Fatalf("checkpoint task does not use mapstructure field names: %#v", task)
	}
	if _, ok := task["TaskName"]; ok {
		t.Fatalf("checkpoint task still uses Go field name: %#v", task)
	}
}
