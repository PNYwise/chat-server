package internal

import (
	"database/sql"
	"log"
	"time"
)

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) IMessageRepository {
	return &messageRepository{db}
}

// Create implements IMessageRepository.
func (m *messageRepository) Create(message *Message) error {
	now := time.Now()
	query := `INSERT INTO messages (from_id, to_id, content, created_at) VALUES (?,?,?,?)`
	result, err := m.db.Exec(query, message.Form.Id, message.To.Id, message.Content, now)
	if err != nil {
		log.Fatalf("Error executing query: %v", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Fatalf("Error get id: %v", err)
		return err
	}
	message.Id = uint(id)
	message.CreatedAt.Time = now

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
