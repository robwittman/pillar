package grpc

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
	"github.com/robwittman/pillar/internal/service"
)

// NewServer creates a gRPC server. When authSvc is non-nil, the auth
// stream interceptor is installed so that agents must present a valid
// Bearer token in metadata.
func NewServer(svc service.AgentService, configSvc service.ConfigService, attrSvc service.AttributeService, logSvc *service.LogService, taskSvc service.TaskService, streams *StreamManager, logger *slog.Logger, authSvc service.AuthService) *grpc.Server {
	var opts []grpc.ServerOption
	if authSvc != nil {
		opts = append(opts, grpc.StreamInterceptor(AuthStreamInterceptor(authSvc)))
	}

	s := grpc.NewServer(opts...)

	var streamService *AgentStreamService
	if streams != nil {
		streamService = NewAgentStreamServiceWithStreams(svc, configSvc, attrSvc, logSvc, taskSvc, streams, logger)
	} else {
		streamService = NewAgentStreamService(svc, configSvc, attrSvc, logSvc, taskSvc, logger)
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
