package server

import (
	"context"
	"strings"

	apiv1 "github.com/HonLaderDev/HonLader-api/pb/honlader/api/v1"
	frame_auth "github.com/HonLaderDev/HonLader/frame/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	authModeExternal = "external"
	authModeBuiltin  = "builtin"
)

var builtinAuthServer frame_auth.AuthServer

// PrepareAuthConfig 将持久化认证配置转换为 Launcher 实际使用的服务地址和令牌。
func (s *dataManagerServer) PrepareAuthConfig(ctx context.Context, req *apiv1.PrepareAuthConfigRequest) (*apiv1.PreparedAuthConfig, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	config, found, err := manager.LoadConfig()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load config: %v", err)
	}
	if !found {
		return nil, status.Error(codes.FailedPrecondition, "auth config is required")
	}
	if normalizeAuthMode(config.AuthMode) == authModeBuiltin {
		if strings.TrimSpace(config.AuthCookie) == "" {
			return nil, status.Error(codes.FailedPrecondition, "auth cookie is required")
		}
		authServer, err := builtinAuthServer.Start(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "start builtin auth server: %v", err)
		}
		return &apiv1.PreparedAuthConfig{AuthServer: authServer, UserToken: config.AuthCookie}, nil
	}
	if strings.TrimSpace(config.AuthServer) == "" || strings.TrimSpace(config.AuthToken) == "" {
		return nil, status.Error(codes.FailedPrecondition, "auth server and auth token are required")
	}
	return &apiv1.PreparedAuthConfig{AuthServer: config.AuthServer, UserToken: config.AuthToken}, nil
}

func normalizeAuthMode(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), authModeBuiltin) {
		return authModeBuiltin
	}
	return authModeExternal
}
