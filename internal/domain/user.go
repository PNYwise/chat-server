package domain

import (
	"database/sql"
)

type User struct {
	Id        uint
	Name      string
	Username  string
	CreatedAt *sql.NullTime
}

type IUserRepository interface {
	Create(user *User) error
	FindByUsername(username string) (*User, error)
}
