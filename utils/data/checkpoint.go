package data

import (
	"fmt"

	"github.com/HonLaderDev/HonLader/define"
)

// SaveCheckpointConfig 保存任务断点配置。
func (m *DataManager) SaveCheckpointConfig(name string, config define.CheckpointConfig) error {
	config.Name = name
	updateConfigTime(&config.CreatedAt, &config.UpdatedAt)
	if err := saveJSON(m.FileSystem(), m.checkpointsDir, config.CreatedAt, config); err != nil {
		return fmt.Errorf("DataManager.SaveCheckpointConfig: %w", err)
	}
	return nil
}

// LoadCheckpointConfig 读取任务断点配置。
func (m *DataManager) LoadCheckpointConfig(name string) (define.CheckpointConfig, bool, error) {
	config, exists, err := loadJSONByName(m.FileSystem(), m.checkpointsDir, name, func(config define.CheckpointConfig) string {
		return config.Name
	})
	if err != nil {
		return define.CheckpointConfig{}, false, fmt.Errorf("DataManager.LoadCheckpointConfig: %w", err)
	}
	return config, exists, nil
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
		if exists && config.Name != "" {
			result[config.Name] = config
		}
	}
	return result, nil
}
