package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

// AuthStreamInterceptor returns a gRPC StreamServerInterceptor that validates
// Bearer tokens from the "authorization" metadata key and injects the resolved
// Principal into the stream context. If org repos are provided and the client
// sends "x-org-id" metadata, it also resolves the org context.
func AuthStreamInterceptor(authSvc service.AuthService, orgRepo domain.OrganizationRepository, membershipRepo domain.MembershipRepository) grpc.StreamServerInterceptor {
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

		principal, oc, err := authSvc.ResolveAPIToken(ss.Context(), token)
		if err != nil {
			return status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx := auth.ContextWithPrincipal(ss.Context(), principal)

		// Use org context from the token if available.
		if oc != nil {
			ctx = auth.ContextWithOrg(ctx, oc)
		} else if orgRepo != nil && membershipRepo != nil {
			// Fall back to x-org-id metadata header.
			ctx, err = resolveGRPCOrg(ctx, md, principal, orgRepo, membershipRepo)
			if err != nil {
				return status.Error(codes.PermissionDenied, err.Error())
			}
		}

		return handler(srv, &authStream{ServerStream: ss, ctx: ctx})
	}
}

func resolveGRPCOrg(ctx context.Context, md metadata.MD, principal *domain.Principal, orgRepo domain.OrganizationRepository, membershipRepo domain.MembershipRepository) (context.Context, error) {
	orgValues := md.Get("x-org-id")

	if len(orgValues) == 0 || orgValues[0] == "" {
		// Auto-select for single-org users.
		memberships, err := membershipRepo.ListByUser(ctx, principal.ID)
		if err != nil || len(memberships) == 0 {
			return ctx, nil // No org context — will fail later if required
		}
		if len(memberships) == 1 {
			org, err := orgRepo.Get(ctx, memberships[0].OrgID)
			if err != nil {
				return ctx, nil
			}
			return auth.ContextWithOrg(ctx, &domain.OrgContext{
				OrgID:   org.ID,
				OrgSlug: org.Slug,
				OrgRole: domain.OrgRole(memberships[0].Role),
			}), nil
		}
		return ctx, nil // Multiple orgs, none selected — will fail if required
	}

	orgID := orgValues[0]
	org, err := orgRepo.Get(ctx, orgID)
	if err != nil {
		org, err = orgRepo.GetBySlug(ctx, orgID)
	}
	if err != nil {
		return nil, domain.ErrOrgNotFound
	}

	membership, err := membershipRepo.GetByOrgAndUser(ctx, org.ID, principal.ID)
	if err != nil {
		return nil, domain.ErrNotAuthorized
	}

	return auth.ContextWithOrg(ctx, &domain.OrgContext{
		OrgID:   org.ID,
		OrgSlug: org.Slug,
		OrgRole: domain.OrgRole(membership.Role),
	}), nil
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
