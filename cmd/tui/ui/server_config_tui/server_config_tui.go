package server_config_tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tui_define "github.com/HonLaderDev/HonLader/cmd/tui/define"
	"github.com/HonLaderDev/HonLader/define"
)

// ServerConfigTUI 负责服务器配置相关的文本交互。
type ServerConfigTUI struct {
	tui tui_define.TextUI
}

// NewServerConfigTUI 创建服务器配置交互实例。
func NewServerConfigTUI(tui tui_define.TextUI) *ServerConfigTUI {
	return &ServerConfigTUI{tui: tui}
}

// Run 显示服务器配置管理页面。
func (t *ServerConfigTUI) Run(ctx context.Context) error {
	for {
		config, result, err := t.SelectServerConfig(ctx)
		if err != nil {
			return err
		}
		if result == ServerConfigSelectReturn {
			return nil
		}
		if err := t.ManageServerConfig(ctx, config); err != nil {
			return err
		}
	}
}

// ServerConfigSelectResult 表示服务器配置选择结果。
type ServerConfigSelectResult int

const (
	ServerConfigSelectCreate ServerConfigSelectResult = iota
	ServerConfigSelectExisting
	ServerConfigSelectReturn
)

// SelectServerConfig 展示已有服务器配置，并允许新建配置。
func (t *ServerConfigTUI) SelectServerConfig(ctx context.Context) (define.ServerConfig, ServerConfigSelectResult, error) {
	configs, err := t.ListServerConfigs()
	if err != nil {
		return define.ServerConfig{}, ServerConfigSelectReturn, err
	}

	options := t.ServerConfigOptions(configs, "返回")
	sel, err := t.tui.Control().Select(ctx, "请选择服务器配置：\n", options)
	if err != nil {
		return define.ServerConfig{}, ServerConfigSelectReturn, err
	}
	if sel == len(options)-1 {
		return define.ServerConfig{}, ServerConfigSelectReturn, nil
	}
	if sel != 0 {
		return configs[sel-1], ServerConfigSelectExisting, nil
	}
	config, err := t.CreateServerConfig(ctx)
	return config, ServerConfigSelectCreate, err
}

// SelectServerConfigForTask 选择任务运行使用的服务器配置，优先提供最近一次使用的配置。
func (t *ServerConfigTUI) SelectServerConfigForTask(ctx context.Context) (define.ServerConfig, ServerConfigSelectResult, error) {
	latest, exists, err := t.tui.DataManager().LoadLatestServerConfig()
	if err != nil {
		return define.ServerConfig{}, ServerConfigSelectReturn, err
	}
	if exists {
		confirm, err := t.tui.Control().Confirm(
			ctx,
			fmt.Sprintf("是否使用最近服务器配置 %s[Y/n]：", t.ServerConfigTitle(latest)),
			true,
		)
		if err != nil {
			return define.ServerConfig{}, ServerConfigSelectReturn, err
		}
		if confirm {
			return latest, ServerConfigSelectExisting, nil
		}
	}

	config, result, err := t.SelectServerConfig(ctx)
	if err != nil || result == ServerConfigSelectReturn {
		return config, result, err
	}
	if err := t.tui.DataManager().SaveLatestServerConfig(config); err != nil {
		return define.ServerConfig{}, ServerConfigSelectReturn, err
	}
	return config, result, nil
}

// ManageServerConfig 管理单个服务器配置。
func (t *ServerConfigTUI) ManageServerConfig(ctx context.Context, config define.ServerConfig) error {
	for {
		sel, err := t.tui.Control().Select(ctx, fmt.Sprintf("请选择服务器配置操作：%s\n", t.ServerConfigTitle(config)), []string{
			"编辑",
			"删除",
			"返回",
		})
		if err != nil {
			return err
		}
		switch sel {
		case 0:
			updated, err := t.EditServerConfig(ctx, config)
			if err != nil {
				return err
			}
			config = updated
		case 1:
			deleted, err := t.DeleteServerConfig(ctx, config)
			if err != nil {
				return err
			}
			if deleted {
				return nil
			}
		case 2:
			return nil
		default:
			return fmt.Errorf("未知服务器配置操作：%d", sel)
		}
	}
}

// CreateServerConfig 交互式创建并保存服务器配置。
func (t *ServerConfigTUI) CreateServerConfig(ctx context.Context) (define.ServerConfig, error) {
	return t.SaveServerConfig(ctx, define.ServerConfig{})
}

// EditServerConfig 交互式编辑并保存服务器配置。
func (t *ServerConfigTUI) EditServerConfig(ctx context.Context, config define.ServerConfig) (define.ServerConfig, error) {
	return t.SaveServerConfig(ctx, config)
}

// DeleteServerConfig 删除服务器配置。
func (t *ServerConfigTUI) DeleteServerConfig(ctx context.Context, config define.ServerConfig) (bool, error) {
	confirm, err := t.tui.Control().Confirm(ctx, fmt.Sprintf("确认删除服务器配置 %s [y/N]：", config.Metadata.Name), false)
	if err != nil {
		return false, err
	}
	if !confirm {
		return false, nil
	}
	deleted, err := t.tui.DataManager().DeleteServerConfig(config.Metadata.Name)
	if err != nil {
		return false, err
	}
	if !deleted {
		t.tui.Control().Print("未找到服务器配置\n")
	}
	return deleted, nil
}

func (t *ServerConfigTUI) SaveServerConfig(ctx context.Context, current define.ServerConfig) (define.ServerConfig, error) {
	serverCode, err := t.PromptServerCode(ctx, current.ServerCode)
	if err != nil {
		return define.ServerConfig{}, err
	}
	serverPassword, err := t.PromptServerPassword(ctx, current.ServerPassword)
	if err != nil {
		return define.ServerConfig{}, err
	}
	serverName, err := t.PromptServerName(ctx, current.Metadata.Name, serverCode)
	if err != nil {
		return define.ServerConfig{}, err
	}

	existing, found, err := t.tui.DataManager().LoadServerConfig(serverName)
	if err != nil {
		return define.ServerConfig{}, err
	}
	if found && existing.Metadata.Name != current.Metadata.Name {
		confirm, err := t.tui.Control().Confirm(ctx, "服务器配置名重复，是否覆盖[y/N]：", false)
		if err != nil {
			return define.ServerConfig{}, err
		}
		if !confirm {
			return existing, nil
		}
		if _, err := t.tui.DataManager().DeleteServerConfig(serverName); err != nil {
			return define.ServerConfig{}, err
		}
	}

	config := current
	config.Metadata.Name = serverName
	config.ServerCode = serverCode
	config.ServerPassword = serverPassword
	if err := t.tui.DataManager().SaveServerConfig(config); err != nil {
		return define.ServerConfig{}, err
	}
	if current.Metadata.Name != "" && current.Metadata.Name != serverName {
		if _, err := t.tui.DataManager().DeleteServerConfig(current.Metadata.Name); err != nil {
			return define.ServerConfig{}, err
		}
	}
	return config, nil
}

func (t *ServerConfigTUI) PromptServerCode(ctx context.Context, current string) (string, error) {
	if strings.TrimSpace(current) != "" {
		return t.tui.PromptRequired(ctx, fmt.Sprintf("请输入服务器号(当前 %s)：", current), current)
	}
	return t.tui.PromptRequired(ctx, "请输入服务器号：")
}

func (t *ServerConfigTUI) PromptServerPassword(ctx context.Context, current string) (string, error) {
	hint := "请输入服务器密码："
	if strings.TrimSpace(current) != "" {
		hint = "请输入服务器密码(留空保持当前)："
	}
	password, err := t.tui.Control().Prompt(ctx, hint)
	if err != nil {
		return "", err
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return current, nil
	}
	return password, nil
}

func (t *ServerConfigTUI) PromptServerName(ctx context.Context, current string, serverCode string) (string, error) {
	defaultName := strings.TrimSpace(current)
	if defaultName == "" {
		defaultName = strings.TrimSpace(serverCode)
	}
	name, err := t.tui.PromptRequired(ctx, fmt.Sprintf("请输入服务器配置名(默认 %s)：", defaultName), defaultName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(name), nil
}

func (t *ServerConfigTUI) ListServerConfigs() ([]define.ServerConfig, error) {
	configMap, err := t.tui.DataManager().ListServerConfigs()
	if err != nil {
		return nil, err
	}
	configs := make([]define.ServerConfig, 0, len(configMap))
	for _, config := range configMap {
		configs = append(configs, config)
	}
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Metadata.Name < configs[j].Metadata.Name
	})
	return configs, nil
}

func (t *ServerConfigTUI) ServerConfigTitle(config define.ServerConfig) string {
	return fmt.Sprintf("%s(%s)", t.tui.Format(config.Metadata.Name), t.tui.Format(config.ServerCode))
}

func (t *ServerConfigTUI) ServerConfigOptions(configs []define.ServerConfig, extraOptions ...string) []string {
	options := make([]string, 0, len(configs)+1+len(extraOptions))
	options = append(options, "新建服务器配置")
	for _, config := range configs {
		options = append(options, t.ServerConfigTitle(config))
	}
	options = append(options, extraOptions...)
	return options
}
