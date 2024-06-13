package configs

import (
	"strconv"
	"time"

	"github.com/PNYwise/chat-server/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type JWTConfig struct {
	config *viper.Viper
}

func NewJWTConfig(config *viper.Viper) *JWTConfig {
	return &JWTConfig{config}
}

func (j *JWTConfig) Generate(user *domain.User) (string, error) {
	duration := j.config.GetInt("jwt.duration")
	key := j.config.GetString("jwt.secret")
	claims := domain.UserClaims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(duration) * time.Hour)),
			ID:        strconv.Itoa(int(user.Id)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}
