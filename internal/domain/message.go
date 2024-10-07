package domain

import (
	"database/sql"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Message struct {
	Id        uint
	Form      *User
	To        *User
	Content   string
	CreatedAt *sql.NullTime
}

type KafkaMessage struct {
	FromId     uint                   `json:"from_id"`
	ToId       uint                   `json:"to_id"`
	ToUsername string                 `json:"to_username"`
	Content    string                 `json:"content"`
	CreatedAt  *timestamppb.Timestamp `json:"created_at"`
}

type IMessageRepository interface {
	Create(message *Message) error
	ReadByUserId(userId uint) (*[]Message, error)
	Delete(ids []uint) error
}
