package service

import (
	"errors"
	"flarrocca/compliant-service/repository/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestReportStolenCard(t *testing.T) {
	type input struct {
		userName   string
		secretCode string
	}

	type output struct {
		response string
		err      error
	}

	type depFields struct {
		userRepositoryMock       *mock.MockUserRepository
		cardRepositoryMock       *mock.MockCardRepository
		stolenCardRepositoryMock *mock.MockStolenCardRepository
	}

	tests := []struct {
		name       string
		input      input
		on         func(*depFields, input)
		assertFunc func(t *testing.T, out output)
	}{
		{
			name: "Success - Cards blocked",
			input: input{
				userName:   "john_doe",
				secretCode: "hashed_secret_123",
			},
			on: func(dep *depFields, in input) {
				dep.userRepositoryMock.EXPECT().GetUser(in.userName).Return(int64(1), "$2a$10$0cdvmI6GCiqRozURednLDOX0wyWHx9HYOOjQhmdFXOSSKYYUC7Oca", nil)
				dep.cardRepositoryMock.EXPECT().GetUserCards(int64(1)).Return([]int64{1, 2}, nil)
				dep.stolenCardRepositoryMock.EXPECT().ReportStolenCards(int64(1), []int64{1, 2}).Return(nil)
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Equal(t, "all the cards linked to the provided user are now blocked. Contact @support-team for more information.", out.response)
				assert.NoError(t, out.err)
			},
		},
		{
			name: "Failure - User not found",
			input: input{
				userName:   "unknown_user",
				secretCode: "some_secret",
			},
			on: func(dep *depFields, in input) {
				dep.userRepositoryMock.EXPECT().GetUser(in.userName).Return(int64(0), "", errors.New("sql: no rows in result set"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Empty(t, out.response)
				assert.EqualError(t, out.err, "user not found")
			},
		},
		{
			name: "Failure - Invalid secret code",
			input: input{
				userName:   "john_doe",
				secretCode: "wrong_secret",
			},
			on: func(dep *depFields, in input) {
				dep.userRepositoryMock.EXPECT().GetUser(in.userName).Return(int64(1), "$2a$10$valid_hashed_secret", nil)
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Empty(t, out.response)
				assert.EqualError(t, out.err, "invalid user name or secret code")
			},
		},
		{
			name: "Failure - Unexpected error in GetUser",
			input: input{
				userName:   "john_doe",
				secretCode: "some_secret",
			},
			on: func(dep *depFields, in input) {
				dep.userRepositoryMock.EXPECT().GetUser(in.userName).Return(int64(0), "", errors.New("database connection error"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Empty(t, out.response)
				assert.EqualError(t, out.err, "database connection error")
			},
		},
		{
			name: "Failure - Report already submitted",
			input: input{
				userName:   "john_doe",
				secretCode: "hashed_secret_123",
			},
			on: func(dep *depFields, in input) {
				dep.userRepositoryMock.EXPECT().GetUser(in.userName).Return(int64(1), "$2a$10$0cdvmI6GCiqRozURednLDOX0wyWHx9HYOOjQhmdFXOSSKYYUC7Oca", nil)
				dep.cardRepositoryMock.EXPECT().GetUserCards(int64(1)).Return([]int64{1, 2}, nil)
				dep.stolenCardRepositoryMock.EXPECT().ReportStolenCards(int64(1), []int64{1, 2}).Return(errors.New("UNIQUE constraint failed: reported_cards.user_id, reported_cards.card_id"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Equal(t, "the report for john_doe has already been submitted.", out.response)
				assert.NoError(t, out.err)
			},
		},
		{
			name: "Failure - Unexpected error in ReportStolenCards",
			input: input{
				userName:   "john_doe",
				secretCode: "hashed_secret_123",
			},
			on: func(dep *depFields, in input) {
				dep.userRepositoryMock.EXPECT().GetUser(in.userName).Return(int64(1), "$2a$10$0cdvmI6GCiqRozURednLDOX0wyWHx9HYOOjQhmdFXOSSKYYUC7Oca", nil)
				dep.cardRepositoryMock.EXPECT().GetUserCards(int64(1)).Return([]int64{1, 2}, nil)
				dep.stolenCardRepositoryMock.EXPECT().ReportStolenCards(int64(1), []int64{1, 2}).Return(errors.New("database timeout error"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Empty(t, out.response)
				assert.EqualError(t, out.err, "database timeout error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepositoryMock := mock.NewMockUserRepository(ctrl)
			cardRepositoryMock := mock.NewMockCardRepository(ctrl)
			stolenCardRepositoryMock := mock.NewMockStolenCardRepository(ctrl)

			tt.on(
				&depFields{
					userRepositoryMock:       userRepositoryMock,
					cardRepositoryMock:       cardRepositoryMock,
					stolenCardRepositoryMock: stolenCardRepositoryMock,
				}, tt.input)

			complianceService := &complianceService{
				userRepository:       userRepositoryMock,
				cardRepository:       cardRepositoryMock,
				stolenCardRepository: stolenCardRepositoryMock,
			}

			response, err := complianceService.ReportStolenCards(tt.input.userName, tt.input.secretCode)

			tt.assertFunc(t, output{response, err})
		})
	}
}

func TestCheckComplianceStatus(t *testing.T) {
	type input struct {
		userID int64
		cardID int64
	}

	type output struct {
		isComplaiance bool
		message       string
		err           error
	}

	type depFields struct {
		stolenCardRepositoryMock *mock.MockStolenCardRepository
		cardRepositoryMock       *mock.MockCardRepository
	}

	tests := []struct {
		name       string
		input      input
		on         func(*depFields, input)
		assertFunc func(t *testing.T, out output)
	}{
		{
			name: "Success - User is compliance",
			input: input{
				userID: 1,
				cardID: 1,
			},
			on: func(dep *depFields, in input) {
				dep.cardRepositoryMock.EXPECT().GetUserCards(in.userID).Return([]int64{1, 2, 3}, nil)
				dep.stolenCardRepositoryMock.EXPECT().IsCardReported(in.userID, in.cardID).Return(false, nil)
			},
			assertFunc: func(t *testing.T, out output) {
				assert.True(t, out.isComplaiance)
				assert.Equal(t, "user is compliance", out.message)
				assert.NoError(t, out.err)
			},
		},
		{
			name: "Success - User is blocked",
			input: input{
				userID: 1,
				cardID: 1,
			},
			on: func(dep *depFields, in input) {
				dep.cardRepositoryMock.EXPECT().GetUserCards(in.userID).Return([]int64{1, 2, 3}, nil)
				dep.stolenCardRepositoryMock.EXPECT().IsCardReported(in.userID, in.cardID).Return(true, nil)
			},
			assertFunc: func(t *testing.T, out output) {
				assert.False(t, out.isComplaiance)
				assert.Equal(t, "user is currently blocked due to reported stolen card/s", out.message)
				assert.NoError(t, out.err)
			},
		},
		{
			name: "Failure - Card does not belong to user",
			input: input{
				userID: 1,
				cardID: 99,
			},
			on: func(dep *depFields, in input) {
				dep.cardRepositoryMock.EXPECT().GetUserCards(in.userID).Return([]int64{1, 2, 3}, nil)
			},
			assertFunc: func(t *testing.T, out output) {
				assert.False(t, out.isComplaiance)
				assert.Equal(t, "the provided card does not belong to the user", out.message)
				assert.NoError(t, out.err)
			},
		},
		{
			name: "Failure - Error retrieving user cards",
			input: input{
				userID: 1,
				cardID: 1,
			},
			on: func(dep *depFields, in input) {
				dep.cardRepositoryMock.EXPECT().GetUserCards(in.userID).Return(nil, errors.New("database error"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.False(t, out.isComplaiance)
				assert.Equal(t, "error retrieving user cards", out.message)
				assert.EqualError(t, out.err, "database error")
			},
		},
		{
			name: "Failure - Unexpected error in IsCardReported",
			input: input{
				userID: 1,
				cardID: 1,
			},
			on: func(dep *depFields, in input) {
				dep.cardRepositoryMock.EXPECT().GetUserCards(in.userID).Return([]int64{1, 2, 3}, nil)
				dep.stolenCardRepositoryMock.EXPECT().IsCardReported(in.userID, in.cardID).Return(false, errors.New("database connection error"))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.False(t, out.isComplaiance)
				assert.Equal(t, "error checking compliance status", out.message)
				assert.EqualError(t, out.err, "database connection error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			stolenCardRepositoryMock := mock.NewMockStolenCardRepository(ctrl)
			cardRepositoryMock := mock.NewMockCardRepository(ctrl)
			dep := &depFields{
				stolenCardRepositoryMock: stolenCardRepositoryMock,
				cardRepositoryMock:       cardRepositoryMock,
			}

			tt.on(dep, tt.input)

			service := &complianceService{
				cardRepository:       cardRepositoryMock,
				stolenCardRepository: stolenCardRepositoryMock,
			}
			blocked, message, err := service.CheckComplianceStatus(tt.input.userID, tt.input.cardID)

			tt.assertFunc(t, output{blocked, message, err})
		})
	}
}
