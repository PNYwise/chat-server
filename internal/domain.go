package internal

import "database/sql"

type User struct {
	Id        uint
	Name      string
	CreatedAt sql.NullTime
}

type Message struct {
	Id        uint
	Form      User
	To        User
	Content   string
	CreatedAt sql.NullTime
}

type IUserRepository interface {
	Create(user *User) error
	Exist(id int) bool
}

type IMessageRepository interface {
	Create(message *Message) error
	ReadByUserId(userId uint) (*[]Message, error)
	Delete(ids []uint) error
}
