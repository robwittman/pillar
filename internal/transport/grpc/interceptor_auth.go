package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/service"
)

// AuthStreamInterceptor returns a gRPC StreamServerInterceptor that validates
// Bearer tokens from the "authorization" metadata key and injects the resolved
// Principal into the stream context.
func AuthStreamInterceptor(authSvc service.AuthService) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization metadata")
		}

		token := values[0]
		if !strings.HasPrefix(token, "Bearer ") {
			return status.Error(codes.Unauthenticated, "invalid authorization format")
		}
		token = strings.TrimPrefix(token, "Bearer ")

		principal, err := authSvc.ResolveAPIToken(ss.Context(), token)
		if err != nil {
			return status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx := auth.ContextWithPrincipal(ss.Context(), principal)
		return handler(srv, &authStream{ServerStream: ss, ctx: ctx})
	}
}

// authStream wraps a grpc.ServerStream to override its Context() with one
// that carries the authenticated Principal.
type authStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authStream) Context() context.Context {
	return s.ctx
}
