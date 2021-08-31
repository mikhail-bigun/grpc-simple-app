package service

import (
	"context"

	"github.com/mikhail-bigun/grpc-app-pcbook/pb/pcbook"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServiceServer server for authentication
type AuthServiceServer struct {
	userStore  UserStore
	jwtManager *JWTManager
}

// NewAuthServiceServer create new auth service server
func NewAuthServiceServer(userStore UserStore, jwtManager *JWTManager) *AuthServiceServer {
	return &AuthServiceServer{
		userStore:  userStore,
		jwtManager: jwtManager,
	}
}

// Login unary rpc for user login
func (server *AuthServiceServer) Login(ctx context.Context, req *pcbook.LoginRequest) (*pcbook.LoginResponse, error) {

	user, err := server.userStore.Find(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password: %v", err)
	}

	token, err := server.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token: %v", err)
	}

	res := &pcbook.LoginResponse{
		AccessToken: token,
	}

	return res, nil
}