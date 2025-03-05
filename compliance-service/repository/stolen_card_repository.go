package repository

import (
	"database/sql"
)

// Run from the /repository folder the following command to generate the mock:
// mockgen -source stolen_card_repository.go -destination mock/stolen_card_repository_mock.go -package mock
type StolenCardRepository interface {
	ReportStolenCards(userID int64, cardIDs []int64) error
	IsCardReported(userID int64, cardID int64) (bool, error)
}

type stolenCardRepository struct {
	db *sql.DB
}

func NewStolenCardRepository(db *sql.DB) StolenCardRepository {
	return &stolenCardRepository{db: db}
}

func (r *stolenCardRepository) ReportStolenCards(userID int64, cardIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO reported_cards (user_id, card_id) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, cardID := range cardIDs {
		_, err := stmt.Exec(userID, cardID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *stolenCardRepository) IsCardReported(userID int64, cardID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM reported_cards WHERE  user_id = ? AND card_id = ?)", userID, cardID).Scan(&exists)
	return exists, err
}
