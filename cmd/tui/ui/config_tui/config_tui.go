package config_tui

import (
	"context"
	"strings"

	tui_define "github.com/HonLaderDev/HonLader/cmd/tui/define"
	"github.com/HonLaderDev/HonLader/define"
)

// ConfigTUI 负责全局配置相关的文本交互。
type ConfigTUI struct {
	tui tui_define.TextUI
}

// NewConfigTUI 创建全局配置交互实例。
func NewConfigTUI(tui tui_define.TextUI) *ConfigTUI {
	return &ConfigTUI{tui: tui}
}

// EnsureAuthConfig 确保验证服务地址和 Token 已配置。
func (t *ConfigTUI) EnsureAuthConfig(ctx context.Context) (define.Config, error) {
	config, _, err := t.tui.DataManager().LoadConfig()
	if err != nil {
		return define.Config{}, err
	}
	if strings.TrimSpace(config.AuthServer) != "" && strings.TrimSpace(config.AuthToken) != "" {
		return config, nil
	}

	t.tui.Control().Print("开始前需要配置验证服务地址和验证服务 Token\n")
	return t.PromptAuthConfig(ctx, config)
}

// PromptAuthConfig 交互式配置验证服务地址和 Token。
func (t *ConfigTUI) PromptAuthConfig(ctx context.Context, current define.Config) (define.Config, error) {
	authServer, err := t.PromptAuthServer(ctx, current.AuthServer)
	if err != nil {
		return define.Config{}, err
	}
	authToken, err := t.PromptAuthToken(ctx, current.AuthToken)
	if err != nil {
		return define.Config{}, err
	}

	config := current
	config.AuthServer = authServer
	config.AuthToken = authToken
	if err := t.tui.DataManager().SaveConfig(config); err != nil {
		return define.Config{}, err
	}
	return config, nil
}

// PromptAuthServer 读取验证服务地址。
func (t *ConfigTUI) PromptAuthServer(ctx context.Context, current string) (string, error) {
	if strings.TrimSpace(current) != "" {
		return t.tui.PromptRequired(ctx, "请输入验证服务地址(留空保持当前)：", current)
	}
	return t.tui.PromptRequired(ctx, "请输入验证服务地址：")
}

// PromptAuthToken 读取验证服务 Token。
func (t *ConfigTUI) PromptAuthToken(ctx context.Context, current string) (string, error) {
	if strings.TrimSpace(current) != "" {
		return t.tui.PromptRequired(ctx, "请输入验证服务 Token(留空保持当前)：", current)
	}
	return t.tui.PromptRequired(ctx, "请输入验证服务 Token：")
}
