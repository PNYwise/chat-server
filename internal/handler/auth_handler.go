package handler

import (
	"context"

	"github.com/PNYwise/chat-server/internal/configs"
	"github.com/PNYwise/chat-server/internal/domain"
	"github.com/PNYwise/chat-server/internal/utils"
	chat_server "github.com/PNYwise/chat-server/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	userRepo  domain.IUserRepository
	jwtConfig *configs.JWTConfig
	chat_server.UnimplementedAuthServer
}

func NewAuthHandler(userRepo domain.IUserRepository, jwtConfig *configs.JWTConfig) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

func (a *AuthHandler) Login(ctx context.Context, request *chat_server.LoginRequest) (*chat_server.Token, error) {
	user, err := a.userRepo.FindByUsername(request.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "wrong username or password")
	}
	if ok := utils.NewBcrypt().VerifyPassword(user.Password, request.GetPassword()); !ok {
		return nil, status.Errorf(codes.Unauthenticated, "wrong username or password")
	}

	token, err := a.jwtConfig.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}
	res := &chat_server.Token{Token: token}
	return res, nil
}

func (a *AuthHandler) Register(ctx context.Context, request *chat_server.RegisterRequest) (*chat_server.Token, error) {
	if ok := a.userRepo.ExistByUsername(request.GetUsername()); ok {
		return nil, status.Errorf(codes.AlreadyExists, "username already exist")
	}
	password, err := utils.NewBcrypt().HashPassword(request.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	user := &domain.User{
		Name:     request.GetName(),
		Username: request.GetUsername(),
		Password: password,
	}
	if err := a.userRepo.Create(user); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	token, err := a.jwtConfig.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &chat_server.Token{Token: token}
	return res, nil
}
