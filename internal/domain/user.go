package domain

import "database/sql"

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
