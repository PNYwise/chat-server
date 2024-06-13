package handler

import (
	"context"

	"github.com/PNYwise/chat-server/internal/domain"
	chat_server "github.com/PNYwise/chat-server/proto"
)

type AuthHandler struct {
	userRepo domain.IUserRepository
	chat_server.UnimplementedAuthServer
}

func NewAuthHandler(userRepo domain.IUserRepository) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
	}
}

func (a *AuthHandler) Login(context.Context, *chat_server.LoginRequest) (*chat_server.Token, error) {
	panic("")
}
func (a *AuthHandler) Register(context.Context, *chat_server.RegisterRequest) (*chat_server.Token, error) {
	panic("")
}
