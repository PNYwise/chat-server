package internal

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
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
		log.Printf("Error executing query: %v", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error get id: %v", err)
		return err
	}
	message.Id = uint(id)
	var createdAt sql.NullTime
	createdAt.Time = now
	createdAt.Valid = true
	message.CreatedAt = &createdAt

	return nil
}

// Delete implements IMessageRepository.
func (m *messageRepository) Delete(ids []uint) error {
	messageids := make([]string, len(ids))
	for i, id := range ids {
		messageids[i] = strconv.Itoa(int(id))
	}
	idStr := strings.Join(messageids, ",")
	query := "DELETE FROM messages WHERE id IN(" + idStr + ")"
	stmt, err := m.db.Prepare(query)
	if err != nil {
		log.Printf("Error preparing query: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return err
	}

	return nil
}

// ReadByUserId implements IMessageRepository.
func (m *messageRepository) ReadByUserId(userId uint) (*[]Message, error) {
	query := `
		SELECT m.id, m.from_id, m.to_id, m.content, m.created_at 
		FROM messages as m 
		WHERE m.to_id = $1`
	rows, err := m.db.Query(query, userId)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var message Message
		var userFrom User
		var userTo User
		var createdAt sql.NullTime
		createdAt.Valid = true
		err := rows.Scan(&message.Id, &userFrom.Id, &userTo.Id, &message.Content, &createdAt.Time)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		message.Form = &userFrom
		message.To = &userTo
		message.CreatedAt = &createdAt
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		return nil, err
	}

	return &messages, nil
}
