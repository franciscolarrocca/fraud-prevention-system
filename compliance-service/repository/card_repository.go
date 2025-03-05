package repository

import (
	"database/sql"
)

// Run from the /repository folder the following command to generate the mock:
// mockgen -source card_repository.go -destination mock/card_repository_mock.go -package mock
type CardRepository interface {
	GetUserCards(userID int64) ([]int64, error)
}

type cardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) CardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) GetUserCards(userID int64) ([]int64, error) {
	rows, err := r.db.Query("SELECT id FROM cards WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cardIDs []int64
	for rows.Next() {
		var cardID int64
		if err := rows.Scan(&cardID); err != nil {
			return nil, err
		}
		cardIDs = append(cardIDs, cardID)
	}

	return cardIDs, nil
}
