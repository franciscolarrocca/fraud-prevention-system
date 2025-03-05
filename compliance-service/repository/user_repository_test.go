package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	type input struct {
		userName string
	}

	type output struct {
		userID       int64
		hashedSecret string
		err          error
	}

	tests := []struct {
		name       string
		input      input
		on         func(dbMock sqlmock.Sqlmock, in input)
		assertFunc func(t *testing.T, out output)
	}{
		{
			name: "Success - User found",
			input: input{
				userName: "john_doe",
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT id, secret_code FROM users WHERE user_name = ?").
					WithArgs(in.userName).
					WillReturnRows(sqlmock.NewRows([]string{"id", "secret_code"}).AddRow(1, "hashed_secret_123"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Equal(t, int64(1), out.userID)
				assert.Equal(t, "hashed_secret_123", out.hashedSecret)
				assert.NoError(t, out.err)
			},
		},
		{
			name: "Failure - User not found",
			input: input{
				userName: "unknown_user",
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT id, secret_code FROM users WHERE user_name = ?").
					WithArgs(in.userName).
					WillReturnError(sql.ErrNoRows)
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Zero(t, out.userID)
				assert.Empty(t, out.hashedSecret)
				assert.EqualError(t, out.err, sql.ErrNoRows.Error())
			},
		},
		{
			name: "Failure - Database error",
			input: input{
				userName: "error_user",
			},
			on: func(dbMock sqlmock.Sqlmock, in input) {
				dbMock.ExpectQuery("SELECT id, secret_code FROM users WHERE user_name = ?").
					WithArgs(in.userName).
					WillReturnError(errors.New("database error"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Zero(t, out.userID)
				assert.Empty(t, out.hashedSecret)
				assert.EqualError(t, out.err, "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, dbMock, _ := sqlmock.New()
			defer db.Close()

			userRepository := NewUserRepository(db)
			tt.on(dbMock, tt.input)

			userID, hashedSecret, err := userRepository.GetUser(tt.input.userName)
			tt.assertFunc(t, output{userID, hashedSecret, err})

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}
