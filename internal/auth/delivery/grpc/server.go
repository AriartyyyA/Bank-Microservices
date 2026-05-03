package grpc

import (
	"context"

	pb "github.com/AriartyyyA/gobank/proto/auth"
)

type TokenValidator interface {
	ValidateToken(token string) (userID, email string, err error)
}

type AuthGRPCServer struct {
	pb.UnimplementedAuthServiceServer
	uc TokenValidator
}

func NewAuthGRPCServer(uc TokenValidator) *AuthGRPCServer {
	return &AuthGRPCServer{uc: uc}
}

func (s *AuthGRPCServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	userID, email, err := s.uc.ValidateToken(req.Token)
	if err != nil {
		return nil, err
	}

	return &pb.ValidateTokenResponse{
		UserId: userID,
		Email:  email,
	}, nil
}
