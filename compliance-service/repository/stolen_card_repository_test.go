package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestReportStolenCards(t *testing.T) {
	type input struct {
		userID  int64
		cardIDs []int64
	}

	tests := []struct {
		name       string
		input      input
		on         func(dbMock sqlmock.Sqlmock, in input)
		assertFunc func(t *testing.T, err error)
	}{
		{
			name: "Success - Cards reported successfully",
			input: input{
				userID:  1,
				cardIDs: []int64{101, 102, 103},
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectBegin()

				stmt := dbMock.ExpectPrepare(`INSERT INTO reported_cards \(user_id, card_id\) VALUES \(\?, \?\)`)

				stmt.ExpectExec().WithArgs(in.userID, in.cardIDs[0]).WillReturnResult(sqlmock.NewResult(1, 1))
				stmt.ExpectExec().WithArgs(in.userID, in.cardIDs[1]).WillReturnResult(sqlmock.NewResult(1, 1))
				stmt.ExpectExec().WithArgs(in.userID, in.cardIDs[2]).WillReturnResult(sqlmock.NewResult(1, 1))

				dbMock.ExpectCommit()
			},
			assertFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "Failure - Begin transaction error",
			input: input{
				userID:  1,
				cardIDs: []int64{101},
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectBegin().WillReturnError(errors.New("failed to begin transaction"))
			},
			assertFunc: func(t *testing.T, err error) {
				assert.EqualError(t, err, "failed to begin transaction")
			},
		},
		{
			name: "Failure - Prepare statement error",
			input: input{
				userID:  1,
				cardIDs: []int64{101},
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectBegin()

				dbMock.ExpectPrepare(`INSERT INTO reported_cards \(user_id, card_id\) VALUES \(\?, \?\)`).
					WillReturnError(errors.New("failed to prepare statement"))
			},
			assertFunc: func(t *testing.T, err error) {
				assert.EqualError(t, err, "failed to prepare statement")
			},
		},
		{
			name: "Failure - Exec error during insert",
			input: input{
				userID:  1,
				cardIDs: []int64{101},
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectBegin()

				stmt := dbMock.ExpectPrepare(`INSERT INTO reported_cards \(user_id, card_id\) VALUES \(\?, \?\)`)

				stmt.ExpectExec().WithArgs(in.userID, in.cardIDs[0]).WillReturnError(errors.New("failed to execute insert"))

				dbMock.ExpectRollback()
			},
			assertFunc: func(t *testing.T, err error) {
				assert.EqualError(t, err, "failed to execute insert")
			},
		},
		{
			name: "Failure - Commit error",
			input: input{
				userID:  1,
				cardIDs: []int64{101, 102},
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectBegin()

				stmt := dbMock.ExpectPrepare(`INSERT INTO reported_cards \(user_id, card_id\) VALUES \(\?, \?\)`)

				stmt.ExpectExec().WithArgs(in.userID, in.cardIDs[0]).WillReturnResult(sqlmock.NewResult(1, 1))
				stmt.ExpectExec().WithArgs(in.userID, in.cardIDs[1]).WillReturnResult(sqlmock.NewResult(1, 1))

				dbMock.ExpectCommit().WillReturnError(errors.New("failed to commit transaction"))
			},
			assertFunc: func(t *testing.T, err error) {
				assert.EqualError(t, err, "failed to commit transaction")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, dbMock, _ := sqlmock.New()
			defer db.Close()

			stolenCardRepository := stolenCardRepository{db: db}
			tt.on(dbMock, tt.input)

			err := stolenCardRepository.ReportStolenCards(tt.input.userID, tt.input.cardIDs)
			tt.assertFunc(t, err)

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestIsCardReported(t *testing.T) {
	type input struct {
		userID int64
		cardID int64
	}

	tests := []struct {
		name       string
		input      input
		on         func(dbMock sqlmock.Sqlmock, in input)
		assertFunc func(t *testing.T, exists bool, err error)
	}{
		{
			name: "Success - Card is reported",
			input: input{
				userID: 1,
				cardID: 101,
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM reported_cards WHERE  user_id = \\? AND card_id = \\?\\)").
					WithArgs(in.userID, in.cardID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			assertFunc: func(t *testing.T, exists bool, err error) {
				assert.True(t, exists)
				assert.NoError(t, err)
			},
		},
		{
			name: "Success - Card is not reported",
			input: input{
				userID: 2,
				cardID: 202,
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM reported_cards WHERE  user_id = \\? AND card_id = \\?\\)").
					WithArgs(in.userID, in.cardID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			assertFunc: func(t *testing.T, exists bool, err error) {
				assert.False(t, exists)
				assert.NoError(t, err)
			},
		},
		{
			name: "Failure - Database error",
			input: input{
				userID: 3,
				cardID: 303,
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM reported_cards WHERE  user_id = \\? AND card_id = \\?\\)").
					WithArgs(in.userID, in.cardID).
					WillReturnError(errors.New("database error"))
			},
			assertFunc: func(t *testing.T, exists bool, err error) {
				assert.Zero(t, exists)
				assert.EqualError(t, err, "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, dbMock, _ := sqlmock.New()
			defer db.Close()

			stolenCardRepository := NewStolenCardRepository(db)
			tt.on(dbMock, tt.input)

			exists, err := stolenCardRepository.IsCardReported(tt.input.userID, tt.input.cardID)
			tt.assertFunc(t, exists, err)

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}
