package internal

import (
	"database/sql"
	"log"
	"time"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) IUserRepository {
	return &userRepository{db}
}

// Create implements IUserRepository.
func (u *userRepository) Create(user *User) error {
	now := time.Now()
	query := `INSERT INTO users (name, created_at) VALUES (?,?)`
	result, err := u.db.Exec(query, user.Name, now)
	if err != nil {
		log.Fatalf("Error executing query: %v", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Fatalf("Error get id: %v", err)
		return err
	}
	user.Id = uint(id)
	user.CreatedAt.Time = now

	return nil
}

// Exist implements IUserRepository.
func (u *userRepository) Exist(id int) bool {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
	var exist bool
	row, err := u.db.Query(query, id)
	if err != nil {
		log.Fatalf("Error executing query: %v", err)
		return false
	}
	for row.Next() {
		if err := row.Scan(&exist); err != nil {
			log.Fatalf("Error Scaning query: %v", err)
			return false
		}
	}
	return exist
}
