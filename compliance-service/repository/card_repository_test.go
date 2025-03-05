package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetUserCards(t *testing.T) {
	type input struct {
		userID int64
	}

	type output struct {
		cardIDs []int64
		err     error
	}

	tests := []struct {
		name       string
		input      input
		on         func(dbMock sqlmock.Sqlmock, in input)
		assertFunc func(t *testing.T, out output)
	}{
		{
			name: "Success - Cards found",
			input: input{
				userID: 1,
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT id FROM cards WHERE user_id = ?").
					WithArgs(in.userID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Equal(t, []int64{1, 2, 3}, out.cardIDs)
				assert.NoError(t, out.err)
			},
		},
		{
			name: "Failure - No cards found",
			input: input{
				userID: 999,
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT id FROM cards WHERE user_id = ?").
					WithArgs(in.userID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Empty(t, out.cardIDs)
				assert.NoError(t, out.err)
			},
		},
		{
			name: "Failure - Database error",
			input: input{
				userID: 1,
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT id FROM cards WHERE user_id = ?").
					WithArgs(in.userID).
					WillReturnError(errors.New("database error"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Empty(t, out.cardIDs)
				assert.EqualError(t, out.err, "database error")
			},
		},
		{
			name: "Failure - Scan error",
			input: input{
				userID: 1,
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT id FROM cards WHERE user_id = ?").
					WithArgs(in.userID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("invalid"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Empty(t, out.cardIDs)
				assert.Error(t, out.err)
				assert.Contains(t, out.err.Error(), "invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, dbMock, _ := sqlmock.New()
			defer db.Close()

			cardRepository := cardRepository{db: db}
			tt.on(dbMock, tt.input)

			cardIDs, err := cardRepository.GetUserCards(tt.input.userID)
			tt.assertFunc(t, output{cardIDs, err})

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}
