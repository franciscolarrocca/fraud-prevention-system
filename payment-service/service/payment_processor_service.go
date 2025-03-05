package service

import (
	"flarrocca/payment-service/repository"
	"fmt"
	"math/rand"
)

// Run from the /service folder the following command to generate the mock:
// mockgen -source payment_processor_service.go -destination mock/payment_processor_service_mock.go -package mock
type PaymentProcessorService interface {
	ProcessPayment(userID int64, cardID int64, amount float64) (string, error)
}

type paymentProcessorService struct {
	complianceRepository repository.ComplianceRepository
}

func NewPaymentProcessorService(complianceRepository repository.ComplianceRepository) PaymentProcessorService {
	return &paymentProcessorService{complianceRepository: complianceRepository}
}

func (p *paymentProcessorService) ProcessPayment(userID int64, cardID int64, amount float64) (string, error) {
	isComplaiance, message := p.complianceRepository.CheckUserComplianceStatus(userID, cardID)
	if !isComplaiance {
		return "", fmt.Errorf("payment denied: %s", message)
	}

	transactionID := fmt.Sprintf("txn_%d", 1000000+rand.Intn(9000000))

	return fmt.Sprintf("payment successful. Transaction ID: %s", transactionID), nil
}
