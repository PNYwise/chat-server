package internal

import (
	"database/sql"
)

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) IMessageRepository {
	return &messageRepository{db}
}

// Create implements IMessageRepository.
func (m *messageRepository) Create(message *Message) error {
	return nil
}

// Delete implements IMessageRepository.
func (m *messageRepository) Delete(ids []uint) error {
	panic("unimplemented")
}

// ReadByUserId implements IMessageRepository.
func (m *messageRepository) ReadByUserId(userId uint) (*[]Message, error) {
	panic("unimplemented")
}
