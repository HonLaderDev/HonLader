package data

import (
	"fmt"

	"github.com/RedLaderDev/RedLader/define"
)

// SaveTaskConfig 保存任务组配置。
func (m *DataManager) SaveTaskConfig(name string, config define.TaskGroupConfig) error {
	config.Name = name
	updateConfigTime(&config.CreatedAt, &config.UpdatedAt)
	if err := saveJSON(m.FileSystem(), m.tasksDir, config.CreatedAt, config); err != nil {
		return fmt.Errorf("DataManager.SaveTaskConfig: %w", err)
	}
	return nil
}

// LoadTaskConfig 读取任务组配置。
func (m *DataManager) LoadTaskConfig(name string) (define.TaskGroupConfig, bool, error) {
	config, exists, err := loadJSONByName(m.FileSystem(), m.tasksDir, name, func(config define.TaskGroupConfig) string {
		return config.Name
	})
	if err != nil {
		return define.TaskGroupConfig{}, false, fmt.Errorf("DataManager.LoadTaskConfig: %w", err)
	}
	return config, exists, nil
}

// ListTaskConfigs 列出所有任务组配置。
func (m *DataManager) ListTaskConfigs() (map[string]define.TaskGroupConfig, error) {
	names, err := listJSONNames(m.FileSystem(), m.tasksDir)
	if err != nil {
		return nil, fmt.Errorf("DataManager.ListTaskConfigs: %w", err)
	}

	result := make(map[string]define.TaskGroupConfig, len(names))
	for _, fileName := range names {
		var config define.TaskGroupConfig
		exists, err := loadJSONFile(m.FileSystem(), m.tasksDir, fileName, &config)
		if err != nil {
			return nil, fmt.Errorf("DataManager.ListTaskConfigs: load %q: %w", fileName, err)
		}
		if exists && config.Name != "" {
			result[config.Name] = config
		}
	}
	return result, nil
}
