package repository

import (
	"database/sql"
)

// Run from the /repository folder the following command to generate the mock:
// mockgen -source user_repository.go -destination mock/user_repository_mock.go -package mock
type UserRepository interface {
	GetUser(userName string) (int64, string, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUser(userName string) (int64, string, error) {
	var userID int64
	var hashedSecret string
	err := r.db.QueryRow("SELECT id, secret_code FROM users WHERE user_name = ?", userName).Scan(&userID, &hashedSecret)
	if err != nil {
		return 0, "", err
	}
	return userID, hashedSecret, nil
}
