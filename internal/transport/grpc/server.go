package grpc

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
	"github.com/robwittman/pillar/internal/service"
)

func NewServer(svc service.AgentService, configSvc service.ConfigService, streams *StreamManager, logger *slog.Logger) *grpc.Server {
	s := grpc.NewServer()

	var streamService *AgentStreamService
	if streams != nil {
		streamService = NewAgentStreamServiceWithStreams(svc, configSvc, streams, logger)
	} else {
		streamService = NewAgentStreamService(svc, configSvc, logger)
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
