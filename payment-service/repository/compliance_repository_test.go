package repository

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUserBlocked(t *testing.T) {
	type input struct {
		userID int64
		cardID int64
	}

	type output struct {
		compliance bool
		message    string
	}

	tests := []struct {
		name       string
		input      input
		mockServer func() *httptest.Server
		assertFunc func(t *testing.T, out output)
	}{
		{
			name: "Success - User is blocked",
			input: input{
				userID: int64(1),
				cardID: int64(1),
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/check_user?user_id=1&card_id=1", r.URL.String())
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"complaiance": false, "message": "user is blocked due to compliance reasons"}`))
				}))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.False(t, out.compliance)
				assert.Equal(t, "user is blocked due to compliance reasons", out.message)
			},
		},
		{
			name: "Success - User is complaiance",
			input: input{
				userID: int64(1),
				cardID: int64(1),
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/check_user?user_id=1&card_id=1", r.URL.String())
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"complaiance": true, "message": "user is complaiance"}`))
				}))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.True(t, out.compliance)
				assert.Equal(t, "user is complaiance", out.message)
			},
		},
		{
			name: "Failure - Error communicating with compliance service",
			input: input{
				userID: int64(1),
				cardID: int64(1),
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.False(t, out.compliance)
				assert.Equal(t, "compliance service returned status code: 500", out.message)
			},
		},
		{
			name: "Failure - Invalid response from compliance service",
			input: input{
				userID: int64(1),
				cardID: int64(1),
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"invalid_json"}`))
				}))
			},
			assertFunc: func(t *testing.T, out output) {
				assert.False(t, out.compliance)
				assert.Equal(t, "error processing compliance response", out.message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.mockServer()
			defer server.Close()

			complianceRepository := &complianceRepository{complianceBaseURL: server.URL}
			blocked, message := complianceRepository.CheckUserComplianceStatus(tt.input.userID, tt.input.cardID)

			tt.assertFunc(t, output{blocked, message})
		})
	}
}
