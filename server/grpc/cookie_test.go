package server

import (
	"context"
	"testing"
	"time"

	apiv1 "github.com/HonLaderDev/HonLader-api/pb/honlader/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCookieInfoRequiresCookie(t *testing.T) {
	_, err := newCookieServer().GetCookieInfo(context.Background(), &apiv1.GetCookieInfoRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("空 Cookie 错误码为 %v，期望 InvalidArgument", status.Code(err))
	}
}

func TestExpiredRegisterSessionCannotContinue(t *testing.T) {
	server := newCookieServer()
	server.sessions["expired"] = registerSession{expiresAt: time.Now().Add(-time.Second)}
	_, err := server.SubmitRegister4399Captcha(context.Background(), &apiv1.SubmitRegister4399CaptchaRequest{SessionId: "expired", CaptchaCode: "ABCD"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("过期会话错误码为 %v，期望 NotFound", status.Code(err))
	}
}
