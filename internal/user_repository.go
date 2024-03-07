package internal

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type userRepository struct {
	db  *pgx.Conn
	ctx context.Context
}

func NewUserRepository(db *pgx.Conn, ctx context.Context) IUserRepository {
	return &userRepository{db, ctx}
}

// Create implements IUserRepository.
func (u *userRepository) Create(user *User) error {
	now := time.Now()
	query := `INSERT INTO users (name, created_at) VALUES ($1,$2) RETURNING id`
	err := u.db.QueryRow(u.ctx, query, user.Name, now).Scan(&user.Id)
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
	row, err := u.db.Query(u.ctx, query, id)
	if err != nil {
		log.Fatalf("Error executing query: %v", err)
		return false
	}

	defer row.Close()

	for row.Next() {
		if err := row.Scan(&exist); err != nil {
			log.Fatalf("Error Scaning query: %v", err)
			return false
		}
	}
	return exist
}
