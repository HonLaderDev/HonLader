package auth_tui

import (
	"context"
	"crypto/rand"
	_ "embed"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	tui_define "github.com/HonLaderDev/HonLader/cmd/tui/define"
	"github.com/HonLaderDev/HonLader/define"
	frame_auth "github.com/HonLaderDev/HonLader/frame/auth"
	auth_define "github.com/HonLaderDev/HonLaderAuth-core/define"
	"github.com/HonLaderDev/HonLaderAuth-neclient/account/com4399"
)

const (
	AuthModeExternal = "external"
	AuthModeBuiltin  = "builtin"
	randomChars      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var builtinAuthServer frame_auth.AuthServer

//go:embed sfz.txt
var realInfoText string

// AuthTUI 负责机器人验证配置。
type AuthTUI struct {
	tui tui_define.TextUI
}

// NewAuthTUI 创建机器人配置 TUI。
func NewAuthTUI(tui tui_define.TextUI) *AuthTUI {
	return &AuthTUI{tui: tui}
}

// Run 显示机器人配置菜单。
func (t *AuthTUI) Run(ctx context.Context) error {
	for {
		config, _, err := t.tui.DataManager().LoadConfig()
		if err != nil {
			return err
		}
		if NormalizeAuthMode(config.AuthMode) == AuthModeBuiltin {
			exit, err := t.RunBuiltinMenu(ctx, config)
			if err != nil {
				return err
			}
			if exit {
				return nil
			}
			continue
		}
		exit, err := t.RunExternalMenu(ctx, config)
		if err != nil {
			return err
		}
		if exit {
			return nil
		}
	}
}

// RunExternalMenu 显示外部验证配置菜单。
func (t *AuthTUI) RunExternalMenu(ctx context.Context, config define.Config) (bool, error) {
	sel, err := t.tui.Control().Select(ctx, "请选择机器人配置：\n", []string{
		"切换内置验证",
		"配置验证服务地址",
		"配置验证服务 Token",
		"返回",
	})
	if err != nil {
		return false, err
	}
	switch sel {
	case 0:
		config.AuthMode = AuthModeBuiltin
	case 1:
		config.AuthServer, err = t.PromptAuthServer(ctx, config.AuthServer)
	case 2:
		config.AuthToken, err = t.PromptAuthToken(ctx, config.AuthToken)
	case 3:
		return true, nil
	default:
		return false, fmt.Errorf("未知机器人配置操作：%d", sel)
	}
	if err != nil {
		return false, err
	}
	return false, t.tui.DataManager().SaveConfig(config)
}

// RunBuiltinMenu 显示内置验证配置菜单。
func (t *AuthTUI) RunBuiltinMenu(ctx context.Context, config define.Config) (bool, error) {
	sel, err := t.tui.Control().Select(ctx, "请选择机器人配置：\n", []string{
		"切换外部验证",
		"获取新4399账号",
		"手动输入 Cookie",
		"显示机器人信息",
		"返回",
	})
	if err != nil {
		return false, err
	}
	switch sel {
	case 0:
		config.AuthMode = AuthModeExternal
	case 1:
		config.AuthCookie, err = t.Register4399(ctx)
	case 2:
		config.AuthCookie, err = t.PromptAuthCookie(ctx, config.AuthCookie)
	case 3:
		if err = t.PrintBotInfo(ctx, config.AuthCookie); err != nil {
			t.tui.Control().Print(fmt.Sprintf("获取机器人信息失败：%v\n", err))
			return false, nil
		}
	case 4:
		return true, nil
	default:
		return false, fmt.Errorf("未知机器人配置操作：%d", sel)
	}
	if err != nil {
		return false, err
	}
	return false, t.tui.DataManager().SaveConfig(config)
}

// PrintBotInfo 显示内置验证 Cookie 对应的机器人信息。
func (t *AuthTUI) PrintBotInfo(ctx context.Context, cookie string) error {
	if strings.TrimSpace(cookie) == "" {
		return fmt.Errorf("请先配置 Cookie")
	}
	t.tui.Control().Print("正在获取机器人信息...\n")
	client, err := (auth_define.DefaultProvider{DisabledAutoUpdateNickname: true}).G79NEClient(ctx, cookie)
	if err != nil {
		return err
	}
	detail, err := client.UserDetail().GetPEUserDetail()
	if err != nil {
		return err
	}
	name := detail.Name
	if name == "" {
		name = detail.Nickname
	}
	message := fmt.Sprintf(
		"机器人信息：\n  用户ID：%d\n  昵称：%s\n  等级：%d\n",
		detail.UserID,
		t.tui.Format(name),
		detail.Level,
	)
	if strings.TrimSpace(detail.Account) != "" {
		message += fmt.Sprintf("  账号：%s\n", t.tui.Format(detail.Account))
	}
	t.tui.Control().Print(message)
	return nil
}

// EnsureAuthConfig 确保构建前验证配置可用。
func (t *AuthTUI) EnsureAuthConfig(ctx context.Context) (define.Config, error) {
	config, _, err := t.tui.DataManager().LoadConfig()
	if err != nil {
		return define.Config{}, err
	}
	mode := NormalizeAuthMode(config.AuthMode)
	if mode == AuthModeBuiltin {
		if strings.TrimSpace(config.AuthCookie) == "" {
			t.tui.Control().Print("开始构建前需要配置内置验证 Cookie\n")
			config.AuthCookie, err = t.PromptAuthCookie(ctx, config.AuthCookie)
			if err != nil {
				return define.Config{}, err
			}
			config.AuthMode = AuthModeBuiltin
			if err := t.tui.DataManager().SaveConfig(config); err != nil {
				return define.Config{}, err
			}
		}
		server, err := builtinAuthServer.Start(ctx)
		if err != nil {
			return define.Config{}, err
		}
		config.AuthServer = server
		config.AuthToken = config.AuthCookie
		return config, nil
	}

	if strings.TrimSpace(config.AuthServer) != "" && strings.TrimSpace(config.AuthToken) != "" {
		config.AuthMode = AuthModeExternal
		return config, nil
	}
	t.tui.Control().Print("开始构建前需要配置外部验证服务地址和验证服务 Token\n")
	config, err = t.PromptExternalAuthConfig(ctx, config)
	if err != nil {
		return define.Config{}, err
	}
	config.AuthMode = AuthModeExternal
	return config, nil
}

// PromptExternalAuthConfig 交互式配置外部验证。
func (t *AuthTUI) PromptExternalAuthConfig(ctx context.Context, current define.Config) (define.Config, error) {
	authServer, err := t.PromptAuthServer(ctx, current.AuthServer)
	if err != nil {
		return define.Config{}, err
	}
	authToken, err := t.PromptAuthToken(ctx, current.AuthToken)
	if err != nil {
		return define.Config{}, err
	}
	current.AuthMode = AuthModeExternal
	current.AuthServer = authServer
	current.AuthToken = authToken
	if err := t.tui.DataManager().SaveConfig(current); err != nil {
		return define.Config{}, err
	}
	return current, nil
}

// PromptAuthServer 读取外部验证服务地址。
func (t *AuthTUI) PromptAuthServer(ctx context.Context, current string) (string, error) {
	if strings.TrimSpace(current) != "" {
		return t.tui.PromptRequired(ctx, "请输入验证服务地址(留空保持当前)：", current)
	}
	return t.tui.PromptRequired(ctx, "请输入验证服务地址：")
}

// PromptAuthToken 读取外部验证服务 Token。
func (t *AuthTUI) PromptAuthToken(ctx context.Context, current string) (string, error) {
	if strings.TrimSpace(current) != "" {
		return t.tui.PromptRequired(ctx, "请输入验证服务 Token(留空保持当前)：", current)
	}
	return t.tui.PromptRequired(ctx, "请输入验证服务 Token：")
}

// PromptAuthCookie 读取内置验证 Cookie。
func (t *AuthTUI) PromptAuthCookie(ctx context.Context, current string) (string, error) {
	if strings.TrimSpace(current) != "" {
		return t.tui.PromptRequired(ctx, "请输入 Cookie(留空保持当前)：", current)
	}
	return t.tui.PromptRequired(ctx, "请输入 Cookie：")
}

// Register4399 自动注册 4399 账号并返回 Cookie。
func (t *AuthTUI) Register4399(ctx context.Context) (string, error) {
	username := "HL" + randomString(15)
	password := "HL" + randomString(15)
	realName, idCard, err := randomRealInfo()
	if err != nil {
		return "", err
	}

	t.tui.Control().Print(fmt.Sprintf("已自动生成账号：%s\n", username))
	t.tui.Control().Print(fmt.Sprintf("已自动生成密码：%s\n", password))
	t.tui.Control().Print("已读取实名信息\n")

	registerCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	client := com4399.Web4399ClientConfig{}.New()
	cookie, err := t.registerWithCaptcha(registerCtx, client, username, password, realName, idCard)
	if err != nil {
		return "", err
	}
	t.tui.Control().Print("4399 注册成功，Cookie 已写入配置\n")
	return cookie, nil
}

func (t *AuthTUI) registerWithCaptcha(ctx context.Context, client *com4399.Web4399Client, username string, password string, realName string, idCard string) (string, error) {
	if err := client.Register(ctx, username, password); err != nil {
		var captchaErr *com4399.WebCaptchaRequiredError
		if !errors.As(err, &captchaErr) {
			return "", err
		}
		for attempt := 0; attempt < 3; attempt++ {
			if captchaErr.CaptchaURL != "" {
				t.tui.Control().Print(fmt.Sprintf("检测到验证码，请访问链接查看：%s\n", captchaErr.CaptchaURL))
			}
			code, err := t.tui.PromptRequired(ctx, "请输入验证码：")
			if err != nil {
				return "", err
			}
			if err = client.AuthCaptcha(ctx, code); err == nil {
				break
			}
			if !errors.As(err, &captchaErr) {
				return "", err
			}
			if attempt == 2 {
				return "", fmt.Errorf("验证码重试次数过多：%w", err)
			}
		}
	}

	if _, err := client.AuthRealName(ctx, realName, idCard); err != nil {
		return "", err
	}
	return com4399.Com4399ClientConfig{}.New().Login(ctx, username, password)
}

// NormalizeAuthMode 规范化验证模式。
func NormalizeAuthMode(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), AuthModeBuiltin) {
		return AuthModeBuiltin
	}
	return AuthModeExternal
}

func randomString(length int) string {
	result := make([]byte, length)
	max := big.NewInt(int64(len(randomChars)))
	for i := 0; i < length; i++ {
		index, _ := rand.Int(rand.Reader, max)
		result[i] = randomChars[index.Int64()]
	}
	return string(result)
}

func randomRealInfo() (string, string, error) {
	lines := make([]string, 0)
	for _, line := range strings.Split(realInfoText, "\n") {
		line := strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return "", "", errors.New("实名数据为空")
	}

	index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lines))))
	parts := strings.SplitN(lines[index.Int64()], ",", 2)
	if len(parts) != 2 {
		return "", "", errors.New("实名数据格式错误，要求：姓名,号码")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}
