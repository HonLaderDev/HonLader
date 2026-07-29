package server

import (
	"context"
	"path/filepath"
	"testing"

	apiv1 "github.com/HonLaderDev/HonLader-api/pb/honlader/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestCreateLauncherCreatesIndependentInstances(t *testing.T) {
	server := New()
	ctx := context.Background()

	first, err := server.CreateLauncher(ctx, &apiv1.CreateLauncherRequest{
		DataDirPath: filepath.Join(t.TempDir(), "first"),
	})
	if err != nil {
		t.Fatalf("创建第一个 Launcher 失败：%v", err)
	}
	second, err := server.CreateLauncher(ctx, &apiv1.CreateLauncherRequest{
		DataDirPath: filepath.Join(t.TempDir(), "second"),
	})
	if err != nil {
		t.Fatalf("创建第二个 Launcher 失败：%v", err)
	}
	if first.GetLauncherId() == second.GetLauncherId() {
		t.Fatalf("并发实例不能复用 ID：%q", first.GetLauncherId())
	}

	result, err := server.ListLaunchers(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("列出 Launcher 失败：%v", err)
	}
	if len(result.GetLauncherIds()) != 2 {
		t.Fatalf("Launcher 数量为 %d，期望 2", len(result.GetLauncherIds()))
	}
}

func TestListSupportedFormatsUsesWaterStructureRegistry(t *testing.T) {
	server := New()
	result, err := server.ListSupportedFormats(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("列出建筑格式失败：%v", err)
	}
	if len(result.GetFormats()) < 30 {
		t.Fatalf("建筑格式数量为 %d，未覆盖 WaterStructure 注册池", len(result.GetFormats()))
	}
	for _, item := range result.GetFormats() {
		if item.GetName() == "" {
			t.Fatal("建筑格式名称不能为空")
		}
	}
}
