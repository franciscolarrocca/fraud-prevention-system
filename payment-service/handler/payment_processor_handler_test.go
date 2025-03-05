package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"flarrocca/payment-service/service/mock"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestProcessPaymentHandler(t *testing.T) {
	type input struct {
		userID int64
		cardID int64
		amount float64
	}

	type depFields struct {
		paymentServiceMock *mock.MockPaymentProcessorService
	}

	tests := []struct {
		name       string
		input      input
		on         func(*depFields, input)
		assertFunc func(t *testing.T, resp *http.Response)
	}{
		{
			name: "Success - Payment processed",
			input: input{
				userID: int64(1),
				cardID: int64(1),
				amount: 100.50,
			},
			on: func(dep *depFields, in input) {
				dep.paymentServiceMock.EXPECT().ProcessPayment(in.userID, in.cardID, 100.50).
					Return("payment successful. Transaction ID: txn_123456", nil)
			},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, `{"message": "payment successful. Transaction ID: txn_123456"}`, string(body))
			},
		},
		{
			name: "Failure - Missing user_id",
			input: input{
				userID: int64(0),
				cardID: int64(1),
				amount: 20.0,
			},
			on: func(dep *depFields, in input) {},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, `{"message": "user id, card id and valid amount are required"}`, string(body))
			},
		},
		{
			name: "Failure - Missing card_id",
			input: input{
				userID: int64(1),
				cardID: int64(0),
				amount: 20.0,
			},
			on: func(dep *depFields, in input) {},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, `{"message": "user id, card id and valid amount are required"}`, string(body))
			},
		},
		{
			name: "Failure - Invalid amount",
			input: input{
				userID: int64(1),
				cardID: int64(1),
				amount: -20.0,
			},
			on: func(dep *depFields, in input) {},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, `{"message": "user id, card id and valid amount are required"}`, string(body))
			},
		},
		{
			name: "Failure - Invalid payment amount",
			input: input{
				userID: int64(1),
				cardID: int64(1),
				amount: -100.50,
			},
			on: func(dep *depFields, in input) {},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, `{"message": "user id, card id and valid amount are required"}`, string(body))
			},
		},
		{
			name: "Failure - User blocked",
			input: input{
				userID: int64(1),
				cardID: int64(1),
				amount: 100.50,
			},
			on: func(dep *depFields, in input) {
				dep.paymentServiceMock.EXPECT().ProcessPayment(in.userID, in.cardID, 100.50).
					Return("", errors.New("payment denied: Suspicious activity detected"))
			},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusForbidden, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, `{"message": "payment denied: Suspicious activity detected"}`, string(body))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			paymentServiceMock := mock.NewMockPaymentProcessorService(ctrl)
			tt.on(&depFields{paymentServiceMock: paymentServiceMock}, tt.input)

			handler := &PaymentProcessorHandler{paymentService: paymentServiceMock}
			app.Post("/process_payment", handler.ProcessPayment)

			payload, _ := json.Marshal(map[string]interface{}{
				"user_id": tt.input.userID,
				"card_id": tt.input.cardID,
				"amount":  tt.input.amount,
			})

			req := httptest.NewRequest(http.MethodPost, "/process_payment", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			tt.assertFunc(t, resp)
		})
	}
}
