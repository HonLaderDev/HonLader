package server

import (
	"context"
	"crypto/rand"
	_ "embed"
	"errors"
	"math/big"
	"strings"
	"sync"
	"time"

	apiv1 "github.com/HonLaderDev/HonLader-api/pb/honlader/api/v1"
	authdefine "github.com/HonLaderDev/HonLaderAuth-core/define"
	"github.com/HonLaderDev/HonLaderAuth-neclient/account/com4399"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const randomChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

//go:embed sfz.txt
var realInfoText string

type registerSession struct {
	client             *com4399.Web4399Client
	username, password string
	expiresAt          time.Time
}

type cookieServer struct {
	apiv1.UnimplementedCookieServiceServer
	mu       sync.Mutex
	sessions map[string]registerSession
}

func newCookieServer() *cookieServer {
	return &cookieServer{sessions: make(map[string]registerSession)}
}

// GetCookieInfo 验证 Cookie，并允许 NEMCClient 为没有昵称的新账号自动初始化昵称。
func (s *cookieServer) GetCookieInfo(ctx context.Context, req *apiv1.GetCookieInfoRequest) (*apiv1.CookieInfo, error) {
	cookie := strings.TrimSpace(req.GetCookie())
	if cookie == "" {
		return nil, status.Error(codes.InvalidArgument, "cookie is required")
	}
	client, err := (authdefine.DefaultProvider{}).G79NEMCClient(ctx, cookie)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "create nemc client: %v", err)
	}
	detail, err := client.UserDetail().GetPEUserDetail()
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "get cookie info: %v", err)
	}
	nickname := detail.Name
	if nickname == "" {
		nickname = detail.Nickname
	}
	return &apiv1.CookieInfo{UserId: detail.UserID, Nickname: nickname, Level: detail.Level, Account: detail.Account}, nil
}

func (s *cookieServer) StartRegister4399(ctx context.Context, _ *emptypb.Empty) (*apiv1.Register4399Response, error) {
	username, password := "HL"+randomString(15), "HL"+randomString(15)
	realName, idCard, err := randomRealInfo()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load real info: %v", err)
	}
	client := com4399.Web4399ClientConfig{}.New()
	// 4399 当前会在首次注册请求中同时校验实名字段；必须在 Register 前注入，
	// 否则页面不会进入注册成功状态，只会返回通用的注册拒绝页面。
	client.SetRegisterIdentity(realName, idCard)
	if err := client.Register(ctx, username, password); err != nil {
		var captcha *com4399.WebCaptchaRequiredError
		if !errors.As(err, &captcha) {
			return nil, status.Errorf(codes.Unavailable, "register 4399: %v", err)
		}
		id := randomString(32)
		s.mu.Lock()
		s.cleanupLocked()
		s.sessions[id] = registerSession{client: client, username: username, password: password, expiresAt: time.Now().Add(10 * time.Minute)}
		s.mu.Unlock()
		return &apiv1.Register4399Response{SessionId: id, Username: username, Password: password, CaptchaUrl: captcha.CaptchaURL}, nil
	}
	return s.finish(ctx, client, username, password)
}

func (s *cookieServer) SubmitRegister4399Captcha(ctx context.Context, req *apiv1.SubmitRegister4399CaptchaRequest) (*apiv1.Register4399Response, error) {
	id, code := strings.TrimSpace(req.GetSessionId()), strings.TrimSpace(req.GetCaptchaCode())
	if id == "" || code == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id and captcha_code are required")
	}
	s.mu.Lock()
	s.cleanupLocked()
	session, ok := s.sessions[id]
	s.mu.Unlock()
	if !ok {
		return nil, status.Error(codes.NotFound, "register session not found or expired")
	}
	if err := session.client.AuthCaptcha(ctx, code); err != nil {
		var captcha *com4399.WebCaptchaRequiredError
		if errors.As(err, &captcha) {
			return &apiv1.Register4399Response{SessionId: id, Username: session.username, Password: session.password, CaptchaUrl: captcha.CaptchaURL}, nil
		}
		return nil, status.Errorf(codes.Unavailable, "submit captcha: %v", err)
	}
	result, err := s.finish(ctx, session.client, session.username, session.password)
	if err == nil {
		s.mu.Lock()
		delete(s.sessions, id)
		s.mu.Unlock()
	}
	return result, err
}

func (s *cookieServer) finish(ctx context.Context, client *com4399.Web4399Client, username, password string) (*apiv1.Register4399Response, error) {
	// 注册产生的网页 Cookie 是后续登录上下文的一部分，不能换用全新的 HTTP Client。
	// 实名已随 Register 一并提交，此处只需复用会话换取游戏 Cookie。
	cookie, err := com4399.Com4399ClientConfig{HTTPClient: client.HTTPClient()}.New().Login(ctx, username, password)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "login 4399: %v", err)
	}
	result := &apiv1.Register4399Response{Username: username, Password: password, Cookie: cookie, Done: true}
	info, infoErr := s.GetCookieInfo(ctx, &apiv1.GetCookieInfoRequest{Cookie: cookie})
	if infoErr != nil {
		result.InfoError = infoErr.Error()
	} else {
		result.BotInfo = info
	}
	return result, nil
}

func (s *cookieServer) cleanupLocked() {
	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.expiresAt) {
			delete(s.sessions, id)
		}
	}
}
func randomString(length int) string {
	result := make([]byte, length)
	max := big.NewInt(int64(len(randomChars)))
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		result[i] = randomChars[n.Int64()]
	}
	return string(result)
}
func randomRealInfo() (string, string, error) {
	lines := make([]string, 0)
	for _, line := range strings.Split(realInfoText, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return "", "", errors.New("real info data is empty")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(lines))))
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(lines[n.Int64()], ",", 2)
	if len(parts) != 2 {
		return "", "", errors.New("real info data format must be name,id_card")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}
