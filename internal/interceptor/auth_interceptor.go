package interceptor

import (
	"context"
	"log"

	"github.com/PNYwise/chat-server/internal/configs"
	"google.golang.org/grpc"
)

type AuthInterceptor struct {
	jwt *configs.JWTConfig
}

func NewAuthInterceptor(jwt *configs.JWTConfig) *AuthInterceptor {
	return &AuthInterceptor{jwt}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Println("--> unary interceptor: ", info.FullMethod)

		// TODO: implement authorization

		return handler(ctx, req)
	}
}

func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		log.Println("--> stream interceptor: ", info.FullMethod)

		// TODO: implement authorization

		return handler(srv, stream)
	}
}
