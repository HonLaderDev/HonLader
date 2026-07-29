package server

import (
	"sync"

	apiv1 "github.com/HonLaderDev/HonLader-api/pb/honlader/api/v1"
	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils/storage"
	"google.golang.org/grpc"
)

const subscriptionBuffer = 64

type Server struct {
	apiv1.UnimplementedLauncherManagerServiceServer
	apiv1.UnimplementedLauncherServiceServer
	apiv1.UnimplementedDataManagerServiceServer
	apiv1.UnimplementedStructureServiceServer

	mu sync.RWMutex

	storage define.Storage

	launchers map[string]*launcherSession

	config  *apiv1.Config
	servers map[string]*apiv1.ServerConfig
	latest  *apiv1.ServerConfig

	nextSubscriberID uint64
	logSubscribers   map[uint64]logSubscriber
	eventSubscribers map[uint64]eventSubscriber
}

// New 创建一个 HonLader API 服务端。
func New() *Server {
	return &Server{
		storage:          storage.NewDefaultStorage(),
		launchers:        make(map[string]*launcherSession),
		servers:          make(map[string]*apiv1.ServerConfig),
		logSubscribers:   make(map[uint64]logSubscriber),
		eventSubscribers: make(map[uint64]eventSubscriber),
	}
}

// Register 将 HonLader API 的所有 gRPC 服务注册到指定 registrar。
func Register(registrar grpc.ServiceRegistrar, srv *Server) {
	apiv1.RegisterLauncherManagerServiceServer(registrar, srv)
	apiv1.RegisterLauncherServiceServer(registrar, srv)
	apiv1.RegisterDataManagerServiceServer(registrar, &dataManagerServer{Server: srv})
	apiv1.RegisterCookieServiceServer(registrar, newCookieServer())
	apiv1.RegisterStructureServiceServer(registrar, srv)
}

func (s *Server) nextIDLocked() uint64 {
	s.nextSubscriberID++
	return s.nextSubscriberID
}
