package domain

import "database/sql"

type Message struct {
	Id        uint
	Form      *User
	To        *User
	Content   string
	CreatedAt *sql.NullTime
}

type IMessageRepository interface {
	Create(message *Message) error
	ReadByUserId(userId uint) (*[]Message, error)
	Delete(ids []uint) error
}
