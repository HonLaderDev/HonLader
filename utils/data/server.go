package data

import (
	"fmt"

	"github.com/RedLaderDev/Fatalder/define"
)

// SaveServerConfig 保存服务器配置。
func (m *DataManager) SaveServerConfig(name string, config define.ServerConfig) error {
	config.Name = name
	updateConfigTime(&config.CreatedAt, &config.UpdatedAt)
	if err := saveJSON(m.FileSystem(), m.serversDir, config.CreatedAt, config); err != nil {
		return fmt.Errorf("DataManager.SaveServerConfig: %w", err)
	}
	return nil
}

// LoadServerConfig 读取服务器配置。
func (m *DataManager) LoadServerConfig(name string) (define.ServerConfig, bool, error) {
	config, exists, err := loadJSONByName(m.FileSystem(), m.serversDir, name, func(config define.ServerConfig) string {
		return config.Name
	})
	if err != nil {
		return define.ServerConfig{}, false, fmt.Errorf("DataManager.LoadServerConfig: %w", err)
	}
	return config, exists, nil
}

// ListServerConfigs 列出所有服务器配置。
func (m *DataManager) ListServerConfigs() (map[string]define.ServerConfig, error) {
	names, err := listJSONNames(m.FileSystem(), m.serversDir)
	if err != nil {
		return nil, fmt.Errorf("DataManager.ListServerConfigs: %w", err)
	}

	result := make(map[string]define.ServerConfig, len(names))
	for _, fileName := range names {
		var config define.ServerConfig
		exists, err := loadJSONFile(m.FileSystem(), m.serversDir, fileName, &config)
		if err != nil {
			return nil, fmt.Errorf("DataManager.ListServerConfigs: load %q: %w", fileName, err)
		}
		if exists && config.Name != "" {
			result[config.Name] = config
		}
	}
	return result, nil
}

// DeleteServerConfig 删除服务器配置。
func (m *DataManager) DeleteServerConfig(name string) (bool, error) {
	deleted, err := deleteJSONByName(m.FileSystem(), m.serversDir, name, func(config define.ServerConfig) string {
		return config.Name
	})
	if err != nil {
		return false, fmt.Errorf("DataManager.DeleteServerConfig: %w", err)
	}
	return deleted, nil
}
