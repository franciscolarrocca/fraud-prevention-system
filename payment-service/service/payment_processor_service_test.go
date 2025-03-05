package service

import (
	"flarrocca/payment-service/repository/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestProcessPayment(t *testing.T) {
	type input struct {
		userID int64
		cardID int64
		amount float64
	}

	type output struct {
		response string
		err      error
	}

	type depFields struct {
		complianceRepositoryMock *mock.MockComplianceRepository
	}

	tests := []struct {
		name       string
		input      input
		on         func(*depFields, input)
		assertFunc func(t *testing.T, out output)
	}{
		{
			name: "Success - Payment processed",
			input: input{
				userID: int64(1),
				cardID: int64(1),
				amount: 100.50,
			},
			on: func(dep *depFields, in input) {
				dep.complianceRepositoryMock.EXPECT().CheckUserComplianceStatus(in.userID, in.cardID).Return(true, "User is complaiance")
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Contains(t, out.response, "payment successful. Transaction ID: txn_")
				assert.NoError(t, out.err)
			},
		},
		{
			name: "Failure - User report",
			input: input{
				userID: int64(1),
				cardID: int64(1),
				amount: 250.00,
			},
			on: func(dep *depFields, in input) {
				dep.complianceRepositoryMock.EXPECT().CheckUserComplianceStatus(in.userID, in.cardID).Return(false, "User is currently blocked due to reported stolen card/s")
			},
			assertFunc: func(t *testing.T, out output) {
				assert.Empty(t, out.response)
				assert.EqualError(t, out.err, "payment denied: User is currently blocked due to reported stolen card/s")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			complianceRepositoryMock := mock.NewMockComplianceRepository(ctrl)
			tt.on(&depFields{complianceRepositoryMock: complianceRepositoryMock}, tt.input)

			service := &paymentProcessorService{complianceRepository: complianceRepositoryMock}
			response, err := service.ProcessPayment(tt.input.userID, tt.input.cardID, tt.input.amount)

			tt.assertFunc(t, output{response, err})
		})
	}
}
