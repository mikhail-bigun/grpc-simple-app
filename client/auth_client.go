package client

import (
	"context"
	"time"

	"github.com/mikhail-bigun/grpc-app-pcbook/pb/pcbook"
	"google.golang.org/grpc"
)

// AuthClient client to call authentication rpcs
type AuthClient struct {
	service  pcbook.AuthServiceClient
	username string
	password string
}

// NewAuthClient create a new auth client
func NewAuthClient(cc *grpc.ClientConn, username, password string) *AuthClient {
	service := pcbook.NewAuthServiceClient(cc)
	return &AuthClient{
		service:  service,
		username: username,
		password: password,
	}
}

// Login make a login rpc request from client
func (client *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pcbook.LoginRequest{
		Username: client.username,
		Password: client.password,
	}

	res, err := client.service.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}
