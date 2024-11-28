package repository

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/PNYwise/chat-server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

// NewUserRepository creates a new instance of IUserRepository for managing user data.
// It requires a context (ctx) for managing request-specific metadata such as timeouts
// or cancellations and a database connection pool (db) for interacting with the database.
//
// Parameters:
//   - ctx: Context for managing request lifecycle, including cancellations and timeouts.
//   - db: A connection pool for PostgreSQL (pgxpool.Pool) to execute queries.
//
// Returns:
//   - An implementation of the domain.IUserRepository interface.
func NewUserRepository(ctx context.Context, db *pgxpool.Pool) domain.IUserRepository {
	return &userRepository{ctx, db}
}

// Create implements IUserRepository.
func (u *userRepository) Create(user *domain.User) error {
	now := time.Now()
	query := `INSERT INTO users (name, username, created_at) VALUES ($1,$2,$3) RETURNING id`
	err := u.db.QueryRow(u.ctx, query, user.Name, user.Username, now).Scan(&user.Id)
	if err != nil {
		log.Fatalf("Error get id: %v", err)
		return err
	}
	var createdAt sql.NullTime
	createdAt.Time = now
	createdAt.Valid = true

	user.CreatedAt = &createdAt

	return nil
}

// FindByUsername fetches a user by username from the database
func (u *userRepository) FindByUsername(username string) (*domain.User, error) {
	query := "SELECT u.id, u.username FROM users u WHERE u.username = $1 LIMIT 1"
	var user domain.User

	row := u.db.QueryRow(u.ctx, query, username)
	if err := row.Scan(&user.Id, &user.Username); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
