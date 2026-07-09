package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/HonLaderDev/HonLader/define"
)

func (t *TextUI) SelectServerConfig(ctx context.Context) (define.ServerConfig, error) {
	configMap, err := t.l.DataManager().ListServerConfigs()
	if err != nil {
		return define.ServerConfig{}, err
	}

	configs := make([]define.ServerConfig, 0, len(configMap))
	for _, config := range configMap {
		configs = append(configs, config)
	}
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Name < configs[j].Name
	})

	options := []string{"新建服务器配置"}
	for _, config := range configs {
		options = append(options, fmt.Sprintf("%s(%s)", config.Name, config.ServerCode))
	}
	sel, err := t.c.Select(ctx, "请选择服务器配置：\n", options)
	if err != nil {
		return define.ServerConfig{}, err
	}
	if sel != 0 {
		return configs[sel-1], nil
	}
	return t.CreateServerConfig(ctx)
}

func (t *TextUI) CreateServerConfig(ctx context.Context) (define.ServerConfig, error) {
	serverCode, err := t.c.Prompt(ctx, "请输入服务器号：")
	if err != nil {
		return define.ServerConfig{}, err
	}
	serverCode = strings.TrimSpace(serverCode)

	serverPassword, err := t.c.Prompt(ctx, "请输入服务器密码：")
	if err != nil {
		return define.ServerConfig{}, err
	}

	serverName, err := t.c.Prompt(ctx, fmt.Sprintf("请输入服务器配置名(默认 %s)：", serverCode))
	if err != nil {
		return define.ServerConfig{}, err
	}
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		serverName = serverCode
	}

	config, found, err := t.l.DataManager().LoadServerConfig(serverName)
	if err != nil {
		return define.ServerConfig{}, err
	}
	if found {
		confirm, err := t.c.Confirm(ctx, "服务器配置名重复，是否覆盖[y/N]：", false)
		if err != nil {
			return define.ServerConfig{}, err
		}
		if !confirm {
			return config, nil
		}
		if _, err := t.l.DataManager().DeleteServerConfig(serverName); err != nil {
			return define.ServerConfig{}, err
		}
	}
	config = define.ServerConfig{
		Name:           serverName,
		ServerCode:     serverCode,
		ServerPassword: serverPassword,
	}
	if err := t.l.DataManager().SaveServerConfig(config); err != nil {
		return define.ServerConfig{}, err
	}
	return config, nil
}
