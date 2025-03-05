package handler

import (
	"errors"
	"flarrocca/compliant-service/service/mock"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestReportStolenCardHandler(t *testing.T) {
	type input struct {
		userName   string
		secretCode string
	}

	type output struct {
		statusCode int
		body       string
	}

	type depFields struct {
		complianceServiceMock *mock.MockComplianceService
	}

	tests := []struct {
		name       string
		input      input
		on         func(*depFields, input)
		assertFunc func(t *testing.T, resp *http.Response)
	}{
		{
			name: "Success - Cards reported",
			input: input{
				userName:   "john_doe",
				secretCode: "secure123",
			},
			on: func(dep *depFields, in input) {
				dep.complianceServiceMock.EXPECT().ReportStolenCards(in.userName, in.secretCode).Return("all the cards linked to the provided user are now blocked. Contact with @support-team for more information.", nil)
			},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, "all the cards linked to the provided user are now blocked. Contact with @support-team for more information.", string(body))
			},
		},
		{
			name: "Failure - Missing user_name or secret_code",
			input: input{
				userName:   "",
				secretCode: "",
			},
			on: func(dep *depFields, in input) {},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, "user name and secret code are required", string(body))
			},
		},
		{
			name: "Failure - Internal service error",
			input: input{
				userName:   "john_doe",
				secretCode: "wrong_secret",
			},
			on: func(dep *depFields, in input) {
				dep.complianceServiceMock.EXPECT().ReportStolenCards(in.userName, in.secretCode).
					Return("", errors.New("internal error"))
			},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, "internal error", string(body))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			complianceServiceMock := mock.NewMockComplianceService(ctrl)
			tt.on(&depFields{complianceServiceMock: complianceServiceMock}, tt.input)

			handler := &ComplianceHandler{complianceService: complianceServiceMock}
			app.Post("/report_cards", handler.ReportStolenCards)

			form := url.Values{}
			form.Set("user_name", tt.input.userName)
			form.Set("secret_code", tt.input.secretCode)
			body := strings.NewReader(form.Encode())

			req := httptest.NewRequest(http.MethodPost, "/report_cards", body)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			tt.assertFunc(t, resp)
		})
	}

}

func TestCheckComplianceStatusHandler(t *testing.T) {
	type input struct {
		userID string
		cardID string
	}

	type depFields struct {
		complianceServiceMock *mock.MockComplianceService
	}

	tests := []struct {
		name       string
		input      input
		on         func(*depFields, input)
		assertFunc func(t *testing.T, resp *http.Response)
	}{
		{
			name: "Success - User is blocked",
			input: input{
				userID: "123",
				cardID: "456",
			},
			on: func(dep *depFields, in input) {
				dep.complianceServiceMock.EXPECT().CheckComplianceStatus(int64(123), int64(456)).Return(false, "user is blocked", nil)
			},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, `{"complaiance": false, "message": "user is blocked"}`, string(body))
			},
		},
		{
			name: "Success - User is not blocked",
			input: input{
				userID: "456",
				cardID: "123",
			},
			on: func(dep *depFields, in input) {
				dep.complianceServiceMock.EXPECT().CheckComplianceStatus(int64(456), int64(123)).Return(true, "user is active", nil)
			},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, `{"complaiance": true, "message": "user is active"}`, string(body))
			},
		},
		{
			name: "Failure - Missing user_id",
			input: input{
				userID: "",
				cardID: "456",
			},
			on: func(dep *depFields, in input) {},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, "user id is required", string(body))
			},
		},
		{
			name: "Failure - Missing card_id",
			input: input{
				userID: "456",
				cardID: "",
			},
			on: func(dep *depFields, in input) {},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, "card id is required", string(body))
			},
		},
		{
			name: "Failure - Invalid user_id format",
			input: input{
				userID: "abc",
				cardID: "456",
			},
			on: func(dep *depFields, in input) {},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(body), "invalid data type for user ID")
			},
		},
		{
			name: "Failure - Invalid card_id format",
			input: input{
				userID: "456",
				cardID: "abc",
			},
			on: func(dep *depFields, in input) {},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(body), "invalid data type for card ID")
			},
		},
		{
			name: "Failure - Internal service error",
			input: input{
				userID: "789",
				cardID: "456",
			},
			on: func(dep *depFields, in input) {
				dep.complianceServiceMock.EXPECT().CheckComplianceStatus(int64(789), int64(456)).Return(false, "error", errors.New("error checking user status"))
			},
			assertFunc: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, `{"complaiance": false, "message": "error checking user status: error"}`, string(body))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			complianceServiceMock := mock.NewMockComplianceService(ctrl)
			tt.on(&depFields{complianceServiceMock: complianceServiceMock}, tt.input)

			handler := &ComplianceHandler{complianceService: complianceServiceMock}
			app.Get("/check_user", handler.CheckComplianceStatus)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/check_user?user_id=%s&card_id=%s", tt.input.userID, tt.input.cardID), nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			tt.assertFunc(t, resp)
		})
	}
}
