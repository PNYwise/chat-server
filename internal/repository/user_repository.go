package repository

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/PNYwise/chat-server/internal/domain"
	"github.com/jackc/pgx/v5"
)

type userRepository struct {
	db  *pgx.Conn
	ctx context.Context
}

func NewUserRepository(db *pgx.Conn, ctx context.Context) domain.IUserRepository {
	return &userRepository{db, ctx}
}

// Create implements IUserRepository.
func (u *userRepository) Create(user *domain.User) error {
	now := time.Now()
	query := `INSERT INTO users (name, username, password, created_at) VALUES ($1,$2,$3,$4) RETURNING id`
	err := u.db.QueryRow(u.ctx, query, user.Name, user.Username, user.Password, now).Scan(&user.Id)
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

// Exist implements IUserRepository.
func (u *userRepository) Exist(id int) bool {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
	var exist bool
	row := u.db.QueryRow(u.ctx, query, id)
	if err := row.Scan(&exist); err != nil {
		log.Fatalf("Error Scaning query: %v", err)
		return false
	}
	return exist
}

func (u *userRepository) ExistByUsername(username string) bool {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	var exist bool
	row := u.db.QueryRow(u.ctx, query, username)
	if err := row.Scan(&exist); err != nil {
		log.Fatalf("Error Scaning query: %v", err)
		return false
	}
	return exist
}

// FindByUsername fetches a user by username from the database
func (u *userRepository) FindByUsername(username string) (*domain.User, error) {
	query := "SELECT u.id, u.username, u.password FROM users u WHERE u.username = $1 LIMIT 1"
	var user domain.User

	row := u.db.QueryRow(u.ctx, query, username)
	if err := row.Scan(&user.Id, &user.Username, &user.Password); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
