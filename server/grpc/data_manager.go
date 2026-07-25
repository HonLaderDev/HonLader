package server

import (
	"context"
	"encoding/json"
	"time"

	apiv1 "github.com/HonLaderDev/HonLader-api/pb/honlader/api/v1"
	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/frame/task/build"
	"github.com/HonLaderDev/HonLader/frame/task/export"
	localdata "github.com/HonLaderDev/HonLader/utils/data"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type dataManagerServer struct {
	*Server
}

const taskTypeKey = "task_type"

func (s *Server) dataManager() define.DataManager {
	return localdata.NewDataManager(s.currentStorage())
}

func (s *Server) launcherDataManager(launcherID string) (define.DataManager, error) {
	session, err := s.getLauncher(launcherID)
	if err != nil {
		return nil, err
	}
	return session.launcher.DataManager(), nil
}

func (s *dataManagerServer) SaveConfig(_ context.Context, req *apiv1.SaveConfigRequest) (*emptypb.Empty, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if err := manager.SaveConfig(toDefineConfig(req.GetConfig())); err != nil {
		return nil, status.Errorf(codes.Internal, "save config: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *dataManagerServer) LoadConfig(_ context.Context, req *apiv1.LoadConfigRequest) (*apiv1.LoadConfigResponse, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	config, found, err := manager.LoadConfig()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load config: %v", err)
	}
	return &apiv1.LoadConfigResponse{Config: fromDefineConfig(config), Found: found}, nil
}

func (s *dataManagerServer) SaveServerConfig(_ context.Context, req *apiv1.SaveServerConfigRequest) (*emptypb.Empty, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if err := manager.SaveServerConfig(toDefineServerConfig(req.GetConfig(), "")); err != nil {
		return nil, status.Errorf(codes.Internal, "save server config: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *dataManagerServer) LoadServerConfig(_ context.Context, req *apiv1.LoadServerConfigRequest) (*apiv1.LoadServerConfigResponse, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	config, found, err := manager.LoadServerConfig(req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load server config: %v", err)
	}
	return &apiv1.LoadServerConfigResponse{Config: fromDefineServerConfig(config), Metadata: fromDefineMetadata(config.Metadata), Found: found}, nil
}

func (s *dataManagerServer) ListServerConfigs(_ context.Context, req *apiv1.ListServerConfigsRequest) (*apiv1.ListServerConfigsResponse, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	configs, err := manager.ListServerConfigs()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list server configs: %v", err)
	}
	names := make([]string, 0, len(configs))
	for name := range configs {
		names = append(names, name)
	}
	return &apiv1.ListServerConfigsResponse{Names: names}, nil
}

func (s *dataManagerServer) DeleteServerConfig(_ context.Context, req *apiv1.DeleteServerConfigRequest) (*emptypb.Empty, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if _, err := manager.DeleteServerConfig(req.GetName()); err != nil {
		return nil, status.Errorf(codes.Internal, "delete server config: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *dataManagerServer) SaveLatestServerConfig(_ context.Context, req *apiv1.SaveServerConfigRequest) (*emptypb.Empty, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if err := manager.SaveLatestServerConfig(toDefineServerConfig(req.GetConfig(), "latest")); err != nil {
		return nil, status.Errorf(codes.Internal, "save latest server config: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *dataManagerServer) LoadLatestServerConfig(_ context.Context, req *apiv1.LoadLatestServerConfigRequest) (*apiv1.LoadServerConfigResponse, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	config, found, err := manager.LoadLatestServerConfig()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load latest server config: %v", err)
	}
	return &apiv1.LoadServerConfigResponse{Config: fromDefineServerConfig(config), Metadata: fromDefineMetadata(config.Metadata), Found: found}, nil
}

func (s *dataManagerServer) DeleteLatestServerConfig(_ context.Context, req *apiv1.DeleteLatestServerConfigRequest) (*emptypb.Empty, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if _, err := manager.DeleteLatestServerConfig(); err != nil {
		return nil, status.Errorf(codes.Internal, "delete latest server config: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *dataManagerServer) SaveTaskGroupCheckpoint(_ context.Context, req *apiv1.SaveTaskGroupCheckpointRequest) (*emptypb.Empty, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if err := manager.SaveTaskGroupCheckpoint(req.GetName(), toDefineTaskGroupCheckpoint(req.GetCheckpoint())); err != nil {
		return nil, status.Errorf(codes.Internal, "save task group checkpoint: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *dataManagerServer) LoadTaskGroupCheckpoint(_ context.Context, req *apiv1.LoadTaskGroupCheckpointRequest) (*apiv1.LoadTaskGroupCheckpointResponse, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	checkpoint, metadata, found, err := manager.LoadTaskGroupCheckpoint(req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load task group checkpoint: %v", err)
	}
	return &apiv1.LoadTaskGroupCheckpointResponse{Checkpoint: fromDefineTaskGroupCheckpoint(checkpoint), Metadata: fromDefineMetadata(metadata), Found: found}, nil
}

func (s *dataManagerServer) ListTaskGroupCheckpoints(_ context.Context, req *apiv1.ListTaskGroupCheckpointsRequest) (*apiv1.ListTaskGroupCheckpointsResponse, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	names, err := manager.ListTaskGroupCheckpoints()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list task group checkpoints: %v", err)
	}
	return &apiv1.ListTaskGroupCheckpointsResponse{Names: names}, nil
}

func (s *dataManagerServer) DeleteTaskGroupCheckpoint(_ context.Context, req *apiv1.DeleteTaskGroupCheckpointRequest) (*emptypb.Empty, error) {
	manager, err := s.launcherDataManager(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if _, err := manager.DeleteTaskGroupCheckpoint(req.GetName()); err != nil {
		return nil, status.Errorf(codes.Internal, "delete task group checkpoint: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func toDefineConfig(config *apiv1.Config) define.Config {
	if config == nil {
		return define.Config{}
	}
	return define.Config{
		AuthServer: config.GetAuthServer(),
		AuthToken:  config.GetAuthToken(),
		AuthMode:   config.GetAuthMode(),
		AuthCookie: config.GetAuthCookie(),
	}
}

func fromDefineConfig(config define.Config) *apiv1.Config {
	return &apiv1.Config{
		AuthServer: config.AuthServer,
		AuthToken:  config.AuthToken,
		AuthMode:   config.AuthMode,
		AuthCookie: config.AuthCookie,
	}
}

func toDefineServerConfig(config *apiv1.ServerConfig, name string) define.ServerConfig {
	if config == nil {
		return define.ServerConfig{}
	}
	if name == "" {
		name = config.GetServerCode()
	}
	return define.ServerConfig{
		Metadata:       define.Metadata{Name: name},
		ServerCode:     config.GetServerCode(),
		ServerPassword: config.GetServerPassword(),
	}
}

func fromDefineServerConfig(config define.ServerConfig) *apiv1.ServerConfig {
	return &apiv1.ServerConfig{
		ServerCode:     config.ServerCode,
		ServerPassword: config.ServerPassword,
	}
}

func fromDefineMetadata(metadata define.Metadata) *apiv1.Metadata {
	return &apiv1.Metadata{
		Name:      metadata.Name,
		CreatedAt: timeToTimestamp(metadata.CreatedAt),
		UpdatedAt: timeToTimestamp(metadata.UpdatedAt),
	}
}

func toDefineTaskGroupCheckpoint(checkpoint *apiv1.TaskGroupCheckpoint) define.TaskGroupCheckpoint {
	if checkpoint == nil {
		return define.TaskGroupCheckpoint{}
	}
	taskConfigs := make([]define.TaskConfig, 0, len(checkpoint.GetTaskConfigs()))
	for _, config := range checkpoint.GetTaskConfigs() {
		taskConfigs = append(taskConfigs, taskConfigToMap(config))
	}
	taskCheckpoints := make([]define.TaskCheckpoint, 0, len(checkpoint.GetTaskCheckpoints()))
	for _, checkpoint := range checkpoint.GetTaskCheckpoints() {
		taskCheckpoints = append(taskCheckpoints, taskCheckpointToMap(checkpoint))
	}
	return define.TaskGroupCheckpoint{
		TaskConfigs:      taskConfigs,
		TaskCheckpoints:  taskCheckpoints,
		Server:           toDefineServerConfig(checkpoint.GetServer(), ""),
		CurrentTaskIndex: int(checkpoint.GetCurrentTaskIndex()),
	}
}

func fromDefineTaskGroupCheckpoint(checkpoint define.TaskGroupCheckpoint) *apiv1.TaskGroupCheckpoint {
	taskConfigs := make([]*apiv1.TaskConfig, 0, len(checkpoint.TaskConfigs))
	for _, config := range checkpoint.TaskConfigs {
		taskConfigs = append(taskConfigs, mapToTaskConfig(config))
	}
	taskCheckpoints := make([]*apiv1.TaskCheckpoint, 0, len(checkpoint.TaskCheckpoints))
	for _, checkpoint := range checkpoint.TaskCheckpoints {
		taskCheckpoints = append(taskCheckpoints, mapToTaskCheckpoint(checkpoint))
	}
	return &apiv1.TaskGroupCheckpoint{
		TaskConfigs:      taskConfigs,
		TaskCheckpoints:  taskCheckpoints,
		Server:           fromDefineServerConfig(checkpoint.Server),
		CurrentTaskIndex: int32(checkpoint.CurrentTaskIndex),
	}
}

func mapToTaskConfig(values map[string]any) *apiv1.TaskConfig {
	if typed, ok := values[taskTypeKey].(string); ok {
		payload := nestedTaskData(values)
		switch typed {
		case build.Name:
			result := new(apiv1.BuildTaskConfig)
			unmarshalMap(payload, result)
			return &apiv1.TaskConfig{Config: &apiv1.TaskConfig_Build{Build: result}}
		case export.Name:
			result := new(apiv1.ExportTaskConfig)
			unmarshalMap(payload, result)
			return &apiv1.TaskConfig{Config: &apiv1.TaskConfig_Export{Export: result}}
		}
	}
	data, _ := json.Marshal(values)
	result := new(apiv1.TaskConfig)
	_ = protojson.Unmarshal(data, result)
	return result
}

func mapToTaskCheckpoint(values map[string]any) *apiv1.TaskCheckpoint {
	if typed, ok := values[taskTypeKey].(string); ok {
		payload := nestedTaskData(values)
		switch typed {
		case build.Name:
			result := new(apiv1.BuildCheckpoint)
			unmarshalMap(payload, result)
			return &apiv1.TaskCheckpoint{Checkpoint: &apiv1.TaskCheckpoint_Build{Build: result}}
		case export.Name:
			result := new(apiv1.ExportCheckpoint)
			unmarshalMap(payload, result)
			return &apiv1.TaskCheckpoint{Checkpoint: &apiv1.TaskCheckpoint_Export{Export: result}}
		}
	}
	data, _ := json.Marshal(values)
	result := new(apiv1.TaskCheckpoint)
	_ = protojson.Unmarshal(data, result)
	return result
}

func nestedTaskData(values map[string]any) map[string]any {
	if values == nil {
		return map[string]any{}
	}
	if nested, ok := values["config"].(map[string]any); ok {
		return nested
	}
	if nested, ok := values["checkpoint"].(map[string]any); ok {
		return nested
	}
	return values
}

func taskConfigToMap(config *apiv1.TaskConfig) map[string]any {
	switch typed := config.GetConfig().(type) {
	case *apiv1.TaskConfig_Build:
		return typedTaskMap(build.Name, "config", typed.Build)
	case *apiv1.TaskConfig_Export:
		return typedTaskMap(export.Name, "config", typed.Export)
	default:
		return map[string]any{}
	}
}

func taskCheckpointToMap(checkpoint *apiv1.TaskCheckpoint) map[string]any {
	switch typed := checkpoint.GetCheckpoint().(type) {
	case *apiv1.TaskCheckpoint_Build:
		return typedTaskMap(build.Name, "checkpoint", typed.Build)
	case *apiv1.TaskCheckpoint_Export:
		return typedTaskMap(export.Name, "checkpoint", typed.Export)
	default:
		return map[string]any{}
	}
}

func typedTaskMap(taskType string, payloadKey string, msg proto.Message) map[string]any {
	return map[string]any{
		taskTypeKey: taskType,
		payloadKey:  protoToMap(msg),
	}
}

func unmarshalMap(values map[string]any, msg proto.Message) {
	data, _ := json.Marshal(values)
	_ = protojson.Unmarshal(data, msg)
}

func protoToMap(msg proto.Message) map[string]any {
	if msg == nil {
		return map[string]any{}
	}
	data, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(msg)
	if err != nil {
		return map[string]any{}
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return map[string]any{}
	}
	return result
}

func timestampToTime(value *timestamppb.Timestamp) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.AsTime()
}

func timeToTimestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value)
}
