package repository

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PNYwise/chat-server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type messageRepository struct {
	db  *pgxpool.Pool
	ctx context.Context
}

// NewMessageRepository creates a new instance of IMessageRepository for managing message data.
// This function initializes the repository with the provided database connection pool and context.
//
// Parameters:
//   - ctx: Context for managing the lifecycle of requests, including timeouts and cancellations.
//   - db: A pgxpool.Pool object representing the PostgreSQL connection pool.
//
// Returns:
//   - An implementation of the domain.IMessageRepository interface.
func NewMessageRepository(ctx context.Context, db *pgxpool.Pool) domain.IMessageRepository {
	return &messageRepository{db, ctx}
}

// Create implements IMessageRepository.
func (m *messageRepository) Create(message *domain.Message) error {
	now := time.Now()
	query := `INSERT INTO messages (from_id, to_id, content, created_at) VALUES ($1,$2,$3,$4) RETURNING id`

	err := m.db.QueryRow(m.ctx, query, message.Form.Id, message.To.Id, message.Content, now).Scan(&message.ID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return err
	}
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

	if _, err := m.db.Exec(m.ctx, query); err != nil {
		log.Fatalf("error executing query: %v", err)
		return err
	}
	return nil
}

// ReadByUserId retrieves messages for a specific user by their ID.
//
// Parameters:
//   - userID: The ID of the user whose messages should be retrieved.
//
// Returns:
//   - A pointer to a slice of domain.Message, containing the user's messages.
//   - An error if something goes wrong during the query.
func (m *messageRepository) ReadByUserId(userID uint) (*[]domain.Message, error) {
	query := `
		SELECT 
			m.id,
			u_from.id,
			u_from.username,
			u_to.id,
			u_to.username, 
			m.content, 
			m.created_at 
		FROM messages as m
		LEFT JOIN users u_from on u_from.id = m.from_id
		LEFT JOIN users u_to on u_to.id = m.to_id
		WHERE m.to_id = $1`
	rows, err := m.db.Query(m.ctx, query, userID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message

	for rows.Next() {
		var message domain.Message
		var userFrom domain.User
		var userTo domain.User
		var createdAt sql.NullTime
		createdAt.Valid = true
		err := rows.Scan(&message.ID, &userFrom.Id, &userFrom.Username, &userTo.Id, &userTo.Username, &message.Content, &createdAt.Time)
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
