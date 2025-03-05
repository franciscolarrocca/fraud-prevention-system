package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type ComplianceResponse struct {
	IsComplaiance bool   `json:"complaiance"`
	Message       string `json:"message"`
}

// Run from the /repository folder the following command to generate the mock:
// mockgen -source compliance_repository.go -destination mock/compliance_repository_mock.go -package mock
type ComplianceRepository interface {
	CheckUserComplianceStatus(userID int64, cardID int64) (bool, string)
}

type complianceRepository struct {
	complianceBaseURL string
}

func NewComplianceRepository() ComplianceRepository {
	complianceBaseURL := os.Getenv("COMPLIANCE_SERVICE_URL")
	if complianceBaseURL == "" {
		complianceBaseURL = "http://localhost:8080"
	}
	return &complianceRepository{
		complianceBaseURL: complianceBaseURL,
	}
}

func (c *complianceRepository) CheckUserComplianceStatus(userID int64, cardID int64) (bool, string) {
	resp, err := http.Get(c.complianceBaseURL + fmt.Sprintf("/check_user?user_id=%s&card_id=%s", strconv.FormatInt(userID, 10), strconv.FormatInt(cardID, 10)))
	if err != nil {
		log.Printf("error calling compliance-service: %v", err)
		return false, "error communicating with compliance service"
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("compliance service returned non-2xx status: %d", resp.StatusCode)
		return false, fmt.Sprintf("compliance service returned status code: %d", resp.StatusCode)
	}

	var result ComplianceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("error decoding compliance response: %v", err)
		return false, "error processing compliance response"
	}

	return result.IsComplaiance, result.Message
}
