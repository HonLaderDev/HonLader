package server

import (
	"context"
	"fmt"
	"io"

	apiv1 "github.com/HonLaderDev/HonLader-api/pb/honlader/api/v1"
	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/frame"
	"github.com/HonLaderDev/HonLader/frame/task/build"
	"github.com/HonLaderDev/HonLader/frame/task/export"
	"github.com/HonLaderDev/HonLader/utils/storage"
	bwo_define "github.com/HonLaderDev/bedrock-world-operator/define"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultLauncherID = "default"

type launcherSession struct {
	id        string
	launcher  define.Launcher
	storage   define.Storage
	logCancel context.CancelFunc
}

type coreLogWriter struct {
	server     *Server
	launcherID string
}

func (w coreLogWriter) Write(p []byte) (int, error) {
	w.server.publishLog(w.launcherID, "core", p)
	return len(p), nil
}

type logSubscriber struct {
	launcherID string
	ch         chan *apiv1.LogEntry
}

type eventSubscriber struct {
	launcherID string
	eventNames map[string]struct{}
	ch         chan *apiv1.LauncherEvent
}

func (s *Server) CreateLauncher(_ context.Context, req *apiv1.CreateLauncherRequest) (*apiv1.LauncherRef, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "create launcher request is required")
	}
	currentStorage, err := s.launcherStorage(req)
	if err != nil {
		return nil, err
	}

	// 每次调用都创建独立实例。旧实现固定复用 default，导致不同用户的
	// 任务、认证配置与日志共享同一个 TaskFrame，无法提供可靠的并发隔离。
	id := uuid.NewString()
	session := &launcherSession{
		id:       id,
		launcher: newRuntimeLauncher(currentStorage),
		storage:  currentStorage,
	}
	s.watchLauncherEvents(session)
	s.startCoreLogWatcher(session)

	s.mu.Lock()
	s.launchers[id] = session
	s.mu.Unlock()
	return &apiv1.LauncherRef{LauncherId: id}, nil
}

func (s *Server) launcherStorage(req *apiv1.CreateLauncherRequest) (define.Storage, error) {
	if req.GetDataDirPath() == "" {
		return s.currentStorage(), nil
	}
	dataDir, err := absPath(req.GetDataDirPath(), "data_dir_path")
	if err != nil {
		return nil, err
	}
	return storage.NewPathStorage(dataDir, dataDir, fmt.Sprintf("%s/tmp", dataDir), fmt.Sprintf("%s/downloads", dataDir)), nil
}

func (s *Server) ListLaunchers(context.Context, *emptypb.Empty) (*apiv1.ListLaunchersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.launchers))
	for id := range s.launchers {
		ids = append(ids, id)
	}
	return &apiv1.ListLaunchersResponse{LauncherIds: ids}, nil
}

func (s *Server) DeleteLauncher(ctx context.Context, req *apiv1.LauncherRef) (*emptypb.Empty, error) {
	if _, err := s.Close(ctx, req); err != nil {
		return nil, err
	}
	id, err := normalizeLauncherID(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	delete(s.launchers, id)
	s.mu.Unlock()
	return &emptypb.Empty{}, nil
}

func (s *Server) Connect(ctx context.Context, req *apiv1.ConnectConfig) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "connect config is required")
	}
	session, err := s.getLauncher(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	authConfig := define.Config{
		AuthServer: req.GetAuthServer(),
		AuthToken:  req.GetAuthToken(),
	}
	if err := session.launcher.DataManager().SaveConfig(authConfig); err != nil {
		return nil, status.Errorf(codes.Internal, "save auth config: %v", err)
	}
	serverConfig := define.ServerConfig{
		Metadata:       define.Metadata{Name: req.GetServerCode()},
		ServerCode:     req.GetServerCode(),
		ServerPassword: req.GetServerPassword(),
	}
	if err := session.launcher.Connect(ctx, serverConfig); err != nil {
		return nil, status.Errorf(codes.Internal, "connect launcher %q: %v", session.id, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) LoadTaskGroup(_ context.Context, req *apiv1.LoadTaskGroupRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "load task group request is required")
	}
	session, err := s.getLauncher(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	tasks, err := tasksFromAPI(req.GetTasks(), nil, session.launcher.TaskFrame())
	if err != nil {
		return nil, err
	}
	if err := session.launcher.LoadTaskGroup(tasks); err != nil {
		return nil, status.Errorf(codes.Internal, "load task group: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) LoadTaskGroupCheckpoint(_ context.Context, req *apiv1.LauncherLoadTaskGroupCheckpointRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "load task group checkpoint request is required")
	}
	session, err := s.getLauncher(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	checkpoint := req.GetCheckpoint()
	if checkpoint == nil {
		return nil, status.Error(codes.InvalidArgument, "checkpoint is required")
	}
	tasks, err := tasksFromAPI(checkpoint.GetTaskConfigs(), checkpoint.GetTaskCheckpoints(), session.launcher.TaskFrame())
	if err != nil {
		return nil, err
	}
	session.launcher.TaskFrame().SetCurrentTaskIndex(int(checkpoint.GetCurrentTaskIndex()))
	if err := session.launcher.LoadTaskGroup(tasks); err != nil {
		return nil, status.Errorf(codes.Internal, "load task group checkpoint: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Start(_ context.Context, req *apiv1.LauncherRef) (*emptypb.Empty, error) {
	session, err := s.getLauncher(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if err := session.launcher.Start(); err != nil {
		return nil, status.Errorf(codes.Internal, "start launcher %q: %v", session.id, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Pause(_ context.Context, req *apiv1.LauncherRef) (*emptypb.Empty, error) {
	session, err := s.getLauncher(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if err := session.launcher.Pause(); err != nil {
		return nil, status.Errorf(codes.Internal, "pause launcher %q: %v", session.id, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Resume(_ context.Context, req *apiv1.LauncherRef) (*emptypb.Empty, error) {
	session, err := s.getLauncher(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if err := session.launcher.Resume(); err != nil {
		return nil, status.Errorf(codes.Internal, "resume launcher %q: %v", session.id, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Stop(_ context.Context, req *apiv1.LauncherRef) (*emptypb.Empty, error) {
	session, err := s.getLauncher(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if err := session.launcher.Stop(); err != nil {
		return nil, status.Errorf(codes.Internal, "stop launcher %q: %v", session.id, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Close(_ context.Context, req *apiv1.LauncherRef) (*emptypb.Empty, error) {
	session, err := s.getLauncher(req.GetLauncherId())
	if err != nil {
		return nil, err
	}
	if session.logCancel != nil {
		session.logCancel()
		session.logCancel = nil
	}
	if err := session.launcher.Close(); err != nil {
		return nil, status.Errorf(codes.Internal, "close launcher %q: %v", session.id, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) WatchLog(req *apiv1.WatchLogRequest, stream apiv1.LauncherService_WatchLogServer) error {
	subscriber := s.addLogSubscriber(req.GetLauncherId())
	defer s.removeLogSubscriber(subscriber)
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case entry := <-subscriber.ch:
			if err := stream.Send(entry); err != nil {
				return err
			}
		}
	}
}

func (s *Server) WatchEvents(req *apiv1.WatchEventsRequest, stream apiv1.LauncherService_WatchEventsServer) error {
	subscriber := s.addEventSubscriber(req.GetLauncherId(), req.GetEventNames())
	defer s.removeEventSubscriber(subscriber)
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case event := <-subscriber.ch:
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}

func newRuntimeLauncher(currentStorage define.Storage) define.Launcher {
	taskFrame := frame.TaskFrameConfig{Embedded: true}.New(nil)
	return frame.NewLauncher(taskFrame, currentStorage)
}

func (s *Server) getLauncher(id string) (*launcherSession, error) {
	normalized, err := normalizeLauncherID(id)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	session, ok := s.launchers[normalized]
	s.mu.RUnlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "launcher %q not found", normalized)
	}
	return session, nil
}

func normalizeLauncherID(id string) (string, error) {
	if id == "" {
		return defaultLauncherID, nil
	}
	return id, nil
}

func (s *Server) startCoreLogWatcher(session *launcherSession) {
	ctx, cancel := context.WithCancel(context.Background())
	session.logCancel = cancel
	go func() {
		if err := session.launcher.WatchLog(ctx, coreLogWriter{server: s, launcherID: session.id}); err != nil && ctx.Err() == nil {
			s.publishLog(session.id, "launcher", []byte(err.Error()))
		}
	}()
}

func (s *Server) watchLauncherEvents(session *launcherSession) {
	bus := session.launcher.EventBus()
	subscribe := func(name string) {
		bus.SubscribeAsync(name, func(args ...any) {
			event := &apiv1.LauncherEvent{
				LauncherId: session.id,
				Time:       timestamppb.Now(),
				Name:       name,
			}
			if len(args) > 0 {
				if index, ok := args[0].(int); ok {
					event.TaskIndex = int32(index)
				}
			}
			if len(args) > 1 {
				if err, ok := args[1].(error); ok {
					event.Error = err.Error()
				}
			}
			s.publishEvent(event)
		}, false)
	}
	subscribe(frame.EventNameTaskFrameTaskAdded)
	subscribe(frame.EventNameTaskFrameStart)
	subscribe(frame.EventNameTaskFrameFinish)
	subscribe(frame.EventNameTaskFrameTaskStart)
	subscribe(frame.EventNameTaskFrameTaskFinish)
	subscribe(frame.EventNameTaskFrameTaskFailed)
	subscribe(frame.EventNameTaskFrameTaskCheckpoint)
	subscribe(frame.EventNameTaskFrameNextTask)
	subscribe(frame.EventNameTaskFramePause)
	subscribe(frame.EventNameTaskFrameResume)
	subscribe(frame.EventNameTaskFrameStop)
	subscribe(frame.EventNameTaskFrameClose)
	subscribe(frame.EventNameLauncherCheckpointSaved)
	subscribe(frame.EventNameLauncherCheckpointFailed)
}

func (s *Server) addLogSubscriber(launcherID string) logSubscriber {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextIDLocked()
	subscriber := logSubscriber{launcherID: launcherID, ch: make(chan *apiv1.LogEntry, subscriptionBuffer)}
	s.logSubscribers[id] = subscriber
	return subscriber
}

func (s *Server) removeLogSubscriber(subscriber logSubscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, item := range s.logSubscribers {
		if item.ch == subscriber.ch {
			delete(s.logSubscribers, id)
			close(item.ch)
			return
		}
	}
}

func (s *Server) publishLog(launcherID string, source string, data []byte) {
	entry := &apiv1.LogEntry{
		LauncherId: launcherID,
		Time:       timestamppb.Now(),
		Source:     source,
		Data:       append([]byte(nil), data...),
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, subscriber := range s.logSubscribers {
		if subscriber.launcherID != "" && subscriber.launcherID != launcherID {
			continue
		}
		select {
		case subscriber.ch <- entry:
		default:
		}
	}
}

func (s *Server) addEventSubscriber(launcherID string, names []string) eventSubscriber {
	nameSet := make(map[string]struct{}, len(names))
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextIDLocked()
	subscriber := eventSubscriber{launcherID: launcherID, eventNames: nameSet, ch: make(chan *apiv1.LauncherEvent, subscriptionBuffer)}
	s.eventSubscribers[id] = subscriber
	return subscriber
}

func (s *Server) removeEventSubscriber(subscriber eventSubscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, item := range s.eventSubscribers {
		if item.ch == subscriber.ch {
			delete(s.eventSubscribers, id)
			close(item.ch)
			return
		}
	}
}

func (s *Server) publishEvent(event *apiv1.LauncherEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, subscriber := range s.eventSubscribers {
		if subscriber.launcherID != "" && subscriber.launcherID != event.GetLauncherId() {
			continue
		}
		if len(subscriber.eventNames) > 0 {
			if _, ok := subscriber.eventNames[event.GetName()]; !ok {
				continue
			}
		}
		select {
		case subscriber.ch <- event:
		default:
		}
	}
}

func tasksFromAPI(configs []*apiv1.TaskConfig, checkpoints []*apiv1.TaskCheckpoint, taskFrame define.TaskFrame) ([]define.Task, error) {
	tasks := make([]define.Task, 0, len(configs))
	for i, config := range configs {
		var checkpoint *apiv1.TaskCheckpoint
		if i < len(checkpoints) {
			checkpoint = checkpoints[i]
		}
		task, err := taskFromAPI(config, checkpoint, taskFrame)
		if err != nil {
			return nil, fmt.Errorf("task %d: %w", i, err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func taskFromAPI(config *apiv1.TaskConfig, checkpoint *apiv1.TaskCheckpoint, taskFrame define.TaskFrame) (define.Task, error) {
	if config == nil {
		return nil, fmt.Errorf("task config is required")
	}
	switch typed := config.GetConfig().(type) {
	case *apiv1.TaskConfig_Build:
		task := toBuildTaskConfig(typed.Build).NewTask(taskFrame)
		if checkpoint != nil {
			if buildTask, ok := task.(*build.BuildTask); ok {
				buildTask.BuildTaskCheckpoint.CurrentChunk = int(checkpoint.GetBuild().GetCurrentChunk())
			}
		}
		return task, nil
	case *apiv1.TaskConfig_Export:
		task := toExportTaskConfig(typed.Export).NewTask(taskFrame)
		if checkpoint != nil {
			if exportTask, ok := task.(*export.ExportTask); ok {
				exportTask.ExportTaskCheckpoint.CurrentBatch = int(checkpoint.GetExport().GetCurrentBatch())
			}
		}
		return task, nil
	default:
		return nil, fmt.Errorf("unsupported task config")
	}
}

func toBuildTaskConfig(config *apiv1.BuildTaskConfig) build.BuildTaskConfig {
	if config == nil {
		return build.BuildTaskConfig{}
	}
	result := build.BuildTaskConfig{
		BuildTaskWorldConfig: build.BuildTaskWorldConfig{
			WorldPath:      config.GetWorldPath(),
			WorldStartPos:  toBlockPos(config.GetWorldStartPos()),
			WorldEndPos:    toBlockPos(config.GetWorldEndPos()),
			WorldDimension: define.Dimension(config.GetWorldDimension()),
		},
		BuildTaskAutoConfig: build.BuildTaskAutoConfig{
			EnableAutoCleanBlock:             config.GetEnableAutoCleanBlock(),
			DisableAutoWaitChunkLoad:         config.GetDisableAutoWaitChunkLoad(),
			DisableAutoFillBuildMode:         config.GetDisableAutoFillBuildMode(),
			DisableAutoCleanItem:             config.GetDisableAutoCleanItem(),
			DisableAutoCommandBlocksDisabled: config.GetDisableAutoCommandBlocksDisabled(),
			DisableAutoUpgradeCommandBlock:   config.GetDisableAutoUpgradeCommandBlock(),
			EnableAutoPlaceDenyBlock:         config.GetEnableAutoPlaceDenyBlock(),
			EnableAutoPlaceBorderBlock:       config.GetEnableAutoPlaceBorderBlock(),
			DisableAutoEnterFixMode:          config.GetDisableAutoEnterFixMode(),
		},
		BuildTaskBuildConfig: build.BuildTaskBuildConfig{
			Progress:            config.GetProgress(),
			DisableGameProgress: config.GetDisableGameProgress(),
			IgnoreCommandBlock:  config.GetIgnoreCommandBlock(),
			IgnoreOtherNBTBlock: config.GetIgnoreOtherNbtBlock(),
		},
		BuildTaskAdvConfig: build.BuildTaskAdvConfig{
			EnterFixModeDirectly:    config.GetEnterFixModeDirectly(),
			UseTickingArea:          config.GetUseTickingArea(),
			PreWaitNextChunkLoad:    config.GetPreWaitNextChunkLoad(),
			PreHandleNextChunkGroup: config.GetPreHandleNextChunkGroup(),
		},
		StartPos:  toBlockPos(config.GetStartPos()),
		Dimension: define.Dimension(config.GetDimension()),
	}
	if config.Speed != nil {
		value := int(config.GetSpeed())
		result.Speed = &value
	}
	if config.ChunkGroupSide != nil {
		value := int(config.GetChunkGroupSide())
		result.ChunkGroupSide = &value
	}
	if config.GameProgressRefreshDelay != nil {
		value := config.GetGameProgressRefreshDelay()
		result.GameProgressRefreshDelay = &value
	}
	if config.FixModeTimeout != nil {
		value := config.GetFixModeTimeout()
		result.FixModeTimeout = &value
	}
	if config.GetConsoleWorldPos() != nil {
		value := toBlockPos(config.GetConsoleWorldPos())
		result.ConsoleWorldPos = &value
	}
	return result
}

func toExportTaskConfig(config *apiv1.ExportTaskConfig) export.ExportTaskConfig {
	if config == nil {
		return export.ExportTaskConfig{}
	}
	return export.ExportTaskConfig{
		FilePath:       config.GetFilePath(),
		StartPos:       toBlockPos(config.GetStartPos()),
		EndPos:         toBlockPos(config.GetEndPos()),
		Dimension:      define.Dimension(toDimensionID(config.GetDimension())),
		RequestTimeout: config.GetRequestTimeout().AsDuration(),
		RetryDelay:     config.GetRetryDelay().AsDuration(),
		ChunkBatchSide: int(config.GetChunkBatchSide()),
	}
}

func toBlockPos(pos *apiv1.BlockPos) define.BlockPos {
	if pos == nil {
		return define.BlockPos{}
	}
	return define.BlockPos{int(pos.GetX()), int(pos.GetY()), int(pos.GetZ())}
}

func toDimensionID(value apiv1.Dimension) int32 {
	switch value {
	case apiv1.Dimension_DIMENSION_NETHER:
		return bwo_define.DimensionIDNether
	case apiv1.Dimension_DIMENSION_END:
		return bwo_define.DimensionIDEnd
	default:
		return bwo_define.DimensionIDOverworld
	}
}

func stringPayload(key string, value string) *structpb.Struct {
	payload, _ := structpb.NewStruct(map[string]any{key: value})
	return payload
}

var _ io.Writer = coreLogWriter{}
