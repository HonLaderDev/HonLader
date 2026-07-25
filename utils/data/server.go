package data

import (
	"fmt"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/spf13/afero"
)

const latestServerFileName = "latest_server"

// SaveServerConfig 保存服务器配置。
func (m *DataManager) SaveServerConfig(config define.ServerConfig) error {
	updateConfigTime(&config.Metadata.CreatedAt, &config.Metadata.UpdatedAt)
	if err := saveJSON(m.FileSystem(), m.serversDir, config.Metadata.CreatedAt, config); err != nil {
		return fmt.Errorf("DataManager.SaveServerConfig: %w", err)
	}
	return nil
}

// SaveLatestServerConfig 保存最近一次使用的服务器配置。
func (m *DataManager) SaveLatestServerConfig(config define.ServerConfig) error {
	updateConfigTime(&config.Metadata.CreatedAt, &config.Metadata.UpdatedAt)
	if err := saveJSONFile(m.FileSystem(), m.configDir, latestServerFileName, config); err != nil {
		return fmt.Errorf("DataManager.SaveLatestServerConfig: %w", err)
	}
	return nil
}

// LoadLatestServerConfig 读取最近一次使用的服务器配置。
func (m *DataManager) LoadLatestServerConfig() (define.ServerConfig, bool, error) {
	var config define.ServerConfig
	exists, err := loadJSONFile(m.FileSystem(), m.configDir, latestServerFileName, &config)
	if err != nil {
		return define.ServerConfig{}, false, fmt.Errorf("DataManager.LoadLatestServerConfig: %w", err)
	}
	if !exists || config.Metadata.Name == "" {
		return define.ServerConfig{}, false, nil
	}
	return config, true, nil
}

// DeleteLatestServerConfig 删除最近一次使用的服务器配置。
func (m *DataManager) DeleteLatestServerConfig() (bool, error) {
	filePath, err := jsonFilePath(m.configDir, latestServerFileName)
	if err != nil {
		return false, fmt.Errorf("DataManager.DeleteLatestServerConfig: %w", err)
	}
	exists, err := afero.Exists(m.FileSystem(), filePath)
	if err != nil {
		return false, fmt.Errorf("DataManager.DeleteLatestServerConfig: stat %q: %w", filePath, err)
	}
	if !exists {
		return false, nil
	}
	if err := m.FileSystem().Remove(filePath); err != nil {
		return false, fmt.Errorf("DataManager.DeleteLatestServerConfig: remove %q: %w", filePath, err)
	}
	return true, nil
}

// LoadServerConfig 读取服务器配置。
func (m *DataManager) LoadServerConfig(name string) (define.ServerConfig, bool, error) {
	config, exists, err := loadJSONByName(m.FileSystem(), m.serversDir, name, func(config define.ServerConfig) string {
		return config.Metadata.Name
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
		if exists && config.Metadata.Name != "" {
			result[config.Metadata.Name] = config
		}
	}
	return result, nil
}

// DeleteServerConfig 删除服务器配置。
func (m *DataManager) DeleteServerConfig(name string) (bool, error) {
	deleted, err := deleteJSONByName(m.FileSystem(), m.serversDir, name, func(config define.ServerConfig) string {
		return config.Metadata.Name
	})
	if err != nil {
		return false, fmt.Errorf("DataManager.DeleteServerConfig: %w", err)
	}
	latest, exists, err := m.LoadLatestServerConfig()
	if err != nil {
		return false, fmt.Errorf("DataManager.DeleteServerConfig: load latest: %w", err)
	}
	if exists && latest.Metadata.Name == name {
		if _, err := m.DeleteLatestServerConfig(); err != nil {
			return false, fmt.Errorf("DataManager.DeleteServerConfig: delete latest: %w", err)
		}
	}
	return deleted, nil
}
