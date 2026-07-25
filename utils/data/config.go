package data

import (
	"fmt"
	"path/filepath"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

const globalConfigFileName = "config.yaml"

// SaveConfig 保存全局配置到数据目录的 config.yaml。
func (m *DataManager) SaveConfig(config define.Config) error {
	filePath := filepath.Join(m.dataDir, globalConfigFileName)
	if err := m.FileSystem().MkdirAll(m.dataDir, 0755); err != nil {
		return fmt.Errorf("DataManager.SaveConfig: mkdir %q: %w", m.dataDir, err)
	}

	encoded, err := utils.MarshalMap(config)
	if err != nil {
		return fmt.Errorf("DataManager.SaveConfig: encode %q: %w", filePath, err)
	}
	data, err := yaml.Marshal(encoded)
	if err != nil {
		return fmt.Errorf("DataManager.SaveConfig: marshal %q: %w", filePath, err)
	}

	tmpPath := filePath + ".tmp"
	if err := afero.WriteFile(m.FileSystem(), tmpPath, data, 0644); err != nil {
		return fmt.Errorf("DataManager.SaveConfig: write temp %q: %w", tmpPath, err)
	}
	if err := m.FileSystem().Rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("DataManager.SaveConfig: rename %q to %q: %w", tmpPath, filePath, err)
	}
	return nil
}

// LoadConfig 从数据目录的 config.yaml 读取全局配置。
func (m *DataManager) LoadConfig() (define.Config, bool, error) {
	filePath := filepath.Join(m.dataDir, globalConfigFileName)
	exists, err := afero.Exists(m.FileSystem(), filePath)
	if err != nil {
		return define.Config{}, false, fmt.Errorf("DataManager.LoadConfig: stat %q: %w", filePath, err)
	}
	if !exists {
		return define.Config{}, false, nil
	}

	data, err := afero.ReadFile(m.FileSystem(), filePath)
	if err != nil {
		return define.Config{}, false, fmt.Errorf("DataManager.LoadConfig: read %q: %w", filePath, err)
	}
	var encoded map[string]any
	if err := yaml.Unmarshal(data, &encoded); err != nil {
		return define.Config{}, false, fmt.Errorf("DataManager.LoadConfig: unmarshal %q: %w", filePath, err)
	}
	var config define.Config
	if err := utils.UnmarshalMap(encoded, &config); err != nil {
		return define.Config{}, false, fmt.Errorf("DataManager.LoadConfig: decode %q: %w", filePath, err)
	}
	return config, true, nil
}
