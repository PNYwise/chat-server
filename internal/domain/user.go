package domain

import (
	"database/sql"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
}

type User struct {
	Id        uint
	Name      string
	Username  string
	Password  string
	CreatedAt *sql.NullTime
}

type IUserRepository interface {
	Create(user *User) error
	Exist(id int) bool
}
