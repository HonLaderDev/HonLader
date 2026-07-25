package data

import (
	"encoding/base64"
	"fmt"

	"github.com/HonLaderDev/HonLader/define"
)

// SaveCheckpointConfig 保存任务断点配置。
func (m *DataManager) SaveCheckpointConfig(name string, config define.CheckpointConfig) error {
	existing, exists, err := m.LoadCheckpointConfig(name)
	if err != nil {
		return fmt.Errorf("DataManager.SaveCheckpointConfig: load existing: %w", err)
	}
	if exists {
		config.Metadata.CreatedAt = existing.Metadata.CreatedAt
	}
	config.Metadata.Name = name
	updateConfigTime(&config.Metadata.CreatedAt, &config.Metadata.UpdatedAt)
	if err := m.deleteCheckpointConfigsByName(name); err != nil {
		return fmt.Errorf("DataManager.SaveCheckpointConfig: delete existing: %w", err)
	}
	if err := saveJSONFile(m.FileSystem(), m.checkpointsDir, checkpointFileName(name), config); err != nil {
		return fmt.Errorf("DataManager.SaveCheckpointConfig: %w", err)
	}
	return nil
}

// LoadCheckpointConfig 读取任务断点配置。
func (m *DataManager) LoadCheckpointConfig(name string) (define.CheckpointConfig, bool, error) {
	var config define.CheckpointConfig
	exists, err := loadJSONFile(m.FileSystem(), m.checkpointsDir, checkpointFileName(name), &config)
	if err != nil {
		return define.CheckpointConfig{}, false, fmt.Errorf("DataManager.LoadCheckpointConfig: %w", err)
	}
	if exists {
		normalizeCheckpointTaskInfo(&config)
		return config, true, nil
	}
	return define.CheckpointConfig{}, false, nil
}

// ListCheckpointConfigs 列出所有任务断点配置。
func (m *DataManager) ListCheckpointConfigs() (map[string]define.CheckpointConfig, error) {
	names, err := listJSONNames(m.FileSystem(), m.checkpointsDir)
	if err != nil {
		return nil, fmt.Errorf("DataManager.ListCheckpointConfigs: %w", err)
	}

	result := make(map[string]define.CheckpointConfig, len(names))
	for _, fileName := range names {
		var config define.CheckpointConfig
		exists, err := loadJSONFile(m.FileSystem(), m.checkpointsDir, fileName, &config)
		if err != nil {
			return nil, fmt.Errorf("DataManager.ListCheckpointConfigs: load %q: %w", fileName, err)
		}
		if exists && config.Metadata.Name != "" {
			normalizeCheckpointTaskInfo(&config)
			result[config.Metadata.Name] = config
		}
	}
	return result, nil
}

// DeleteCheckpointConfig 删除任务断点配置。
func (m *DataManager) DeleteCheckpointConfig(name string) (bool, error) {
	_, exists, err := m.LoadCheckpointConfig(name)
	if err != nil {
		return false, fmt.Errorf("DataManager.DeleteCheckpointConfig: load: %w", err)
	}
	if !exists {
		return false, nil
	}
	if err := m.deleteCheckpointConfigsByName(name); err != nil {
		return false, fmt.Errorf("DataManager.DeleteCheckpointConfig: %w", err)
	}
	return true, nil
}

// SaveTaskGroupCheckpoint 保存任务组断点。
func (m *DataManager) SaveTaskGroupCheckpoint(name string, checkpoint define.TaskGroupCheckpoint) error {
	return m.SaveCheckpointConfig(name, checkpointConfigFromTaskGroupCheckpoint(checkpoint))
}

// LoadTaskGroupCheckpoint 读取任务组断点。
func (m *DataManager) LoadTaskGroupCheckpoint(name string) (define.TaskGroupCheckpoint, define.Metadata, bool, error) {
	config, exists, err := m.LoadCheckpointConfig(name)
	if err != nil {
		return define.TaskGroupCheckpoint{}, define.Metadata{}, false, fmt.Errorf("DataManager.LoadTaskGroupCheckpoint: %w", err)
	}
	if !exists {
		return define.TaskGroupCheckpoint{}, define.Metadata{}, false, nil
	}
	metadata := define.Metadata{
		Name:      config.Metadata.Name,
		CreatedAt: config.Metadata.CreatedAt,
		UpdatedAt: config.Metadata.UpdatedAt,
	}
	return taskGroupCheckpointFromCheckpointConfig(config), metadata, true, nil
}

// ListTaskGroupCheckpoints 列出所有任务组断点名称。
func (m *DataManager) ListTaskGroupCheckpoints() ([]string, error) {
	configs, err := m.ListCheckpointConfigs()
	if err != nil {
		return nil, fmt.Errorf("DataManager.ListTaskGroupCheckpoints: %w", err)
	}
	names := make([]string, 0, len(configs))
	for name := range configs {
		names = append(names, name)
	}
	return names, nil
}

// DeleteTaskGroupCheckpoint 删除任务组断点。
func (m *DataManager) DeleteTaskGroupCheckpoint(name string) (bool, error) {
	deleted, err := m.DeleteCheckpointConfig(name)
	if err != nil {
		return false, fmt.Errorf("DataManager.DeleteTaskGroupCheckpoint: %w", err)
	}
	return deleted, nil
}

func (m *DataManager) deleteCheckpointConfigsByName(name string) error {
	names, err := listJSONNames(m.FileSystem(), m.checkpointsDir)
	if err != nil {
		return err
	}
	for _, fileName := range names {
		var config define.CheckpointConfig
		exists, err := loadJSONFile(m.FileSystem(), m.checkpointsDir, fileName, &config)
		if err != nil {
			return fmt.Errorf("load %q: %w", fileName, err)
		}
		if !exists || config.Metadata.Name != name {
			continue
		}
		filePath, err := jsonFilePath(m.checkpointsDir, fileName)
		if err != nil {
			return fmt.Errorf("path %q: %w", fileName, err)
		}
		if err := m.FileSystem().Remove(filePath); err != nil {
			return fmt.Errorf("remove %q: %w", filePath, err)
		}
	}
	return nil
}

func checkpointFileName(name string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(name))
}

func checkpointConfigFromTaskGroupCheckpoint(checkpoint define.TaskGroupCheckpoint) define.CheckpointConfig {
	tasks := make([]define.TaskInfo, 0, len(checkpoint.TaskConfigs))
	for i, config := range checkpoint.TaskConfigs {
		info := define.TaskInfo{Config: config}
		if i < len(checkpoint.TaskCheckpoints) {
			info.Checkpoint = checkpoint.TaskCheckpoints[i]
		}
		tasks = append(tasks, info)
	}
	return define.CheckpointConfig{Tasks: tasks}
}

func taskGroupCheckpointFromCheckpointConfig(config define.CheckpointConfig) define.TaskGroupCheckpoint {
	result := define.TaskGroupCheckpoint{
		TaskConfigs:     make([]define.TaskConfig, 0, len(config.Tasks)),
		TaskCheckpoints: make([]define.TaskCheckpoint, 0, len(config.Tasks)),
	}
	for _, task := range config.Tasks {
		result.TaskConfigs = append(result.TaskConfigs, task.Config)
		result.TaskCheckpoints = append(result.TaskCheckpoints, task.Checkpoint)
	}
	return result
}

func normalizeCheckpointTaskInfo(config *define.CheckpointConfig) {
	for i := range config.Tasks {
		task := &config.Tasks[i]
		if task.TaskName != "" {
			continue
		}
		if len(task.Config) == 0 {
			continue
		}
		if name, ok := task.Config["TaskName"].(string); ok {
			task.TaskName = name
		}
		if configMap, ok := task.Config["Config"].(map[string]any); ok {
			task.Config = configMap
		}
		if checkpointMap, ok := task.Config["Checkpoint"].(map[string]any); ok {
			task.Checkpoint = checkpointMap
		}
	}
}
