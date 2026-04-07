package grpc

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

// GRPCServerConfig holds dependencies for creating a gRPC server.
type GRPCServerConfig struct {
	AgentSvc       service.AgentService
	ConfigSvc      service.ConfigService
	AttrSvc        service.AttributeService
	LogSvc         *service.LogService
	TaskSvc        service.TaskService
	Streams        *StreamManager
	Logger         *slog.Logger
	AuthSvc        service.AuthService
	OrgRepo        domain.OrganizationRepository
	MembershipRepo domain.MembershipRepository
}

// NewServer creates a gRPC server. When AuthSvc is non-nil, the auth
// stream interceptor is installed so that agents must present a valid
// Bearer token in metadata.
func NewServer(cfg GRPCServerConfig) *grpc.Server {
	var opts []grpc.ServerOption
	if cfg.AuthSvc != nil {
		opts = append(opts, grpc.StreamInterceptor(AuthStreamInterceptor(cfg.AuthSvc, cfg.OrgRepo, cfg.MembershipRepo)))
	}

	s := grpc.NewServer(opts...)

	var streamService *AgentStreamService
	if cfg.Streams != nil {
		streamService = NewAgentStreamServiceWithStreams(cfg.AgentSvc, cfg.ConfigSvc, cfg.AttrSvc, cfg.LogSvc, cfg.TaskSvc, cfg.Streams, cfg.Logger)
	} else {
		streamService = NewAgentStreamService(cfg.AgentSvc, cfg.ConfigSvc, cfg.AttrSvc, cfg.LogSvc, cfg.TaskSvc, cfg.Logger)
	}
	pillarv1.RegisterAgentStreamServiceServer(s, streamService)
	reflection.Register(s)

	return s
}

func ListenAndServe(s *grpc.Server, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(lis)
}
