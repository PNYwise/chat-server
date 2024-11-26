package domain

import (
	"database/sql"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Message struct {
	ID        uint
	Form      *User
	To        *User
	Content   string
	CreatedAt *sql.NullTime
}

type KafkaMessage struct {
	FromID       uint                   `json:"from_id"`
	ToID         uint                   `json:"to_id"`
	FromUsername string                 `json:"from_username"`
	ToUsername   string                 `json:"to_username"`
	Content      string                 `json:"content"`
	CreatedAt    *timestamppb.Timestamp `json:"created_at"`
}

type IMessageRepository interface {
	Create(message *Message) error
	ReadByUserId(userId uint) (*[]Message, error)
	Delete(ids []uint) error
}
