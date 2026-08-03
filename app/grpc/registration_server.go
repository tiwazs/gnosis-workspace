package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	workspacev1 "github.com/tiwazs/gnosis-workspace/app/grpc/workspace/v1"
	"github.com/tiwazs/gnosis-workspace/app/services"
)

type RegistrationServer struct {
    workspacev1.UnimplementedRegistrationServiceServer 
    Service *services.WorkspaceService
}

func (server *RegistrationServer) RedeemToken(ctx context.Context, req *workspacev1.RedeemTokenRequest) (*workspacev1.RedeemTokenResponse, error) {
	record, err := server.Service.RedeemToken(req.GetToken())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not redeem token: %v", err)
	}

	return &workspacev1.RedeemTokenResponse{
		WorkspaceId: record.WorkspaceID.String(),
		CreatedBy:   record.CreatedBy,
	}, nil
}