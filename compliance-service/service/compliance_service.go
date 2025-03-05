package service

import (
	"errors"
	"flarrocca/compliant-service/repository"
	"fmt"

	"slices"

	"golang.org/x/crypto/bcrypt"
)

// Run from the /service folder the following command to generate the mock:
// mockgen -source compliance_service.go -destination mock/compliance_service_mock.go -package mock
type ComplianceService interface {
	ReportStolenCards(userName, secretCode string) (string, error)
	CheckComplianceStatus(userID int64, cardID int64) (bool, string, error)
}

type complianceService struct {
	userRepository       repository.UserRepository
	cardRepository       repository.CardRepository
	stolenCardRepository repository.StolenCardRepository
}

func NewComplianceService(userRepository repository.UserRepository, cardRepository repository.CardRepository, stolenCardRepository repository.StolenCardRepository) ComplianceService {
	return &complianceService{
		userRepository:       userRepository,
		cardRepository:       cardRepository,
		stolenCardRepository: stolenCardRepository,
	}
}

func (s *complianceService) ReportStolenCards(userName, secretCode string) (string, error) {
	userID, hashedSecret, err := s.userRepository.GetUser(userName)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return "", errors.New("user not found")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(secretCode)); err != nil {
		return "", errors.New("invalid user name or secret code")
	}

	cardIDs, err := s.cardRepository.GetUserCards(userID)
	if err != nil {
		return "", err
	}

	if len(cardIDs) == 0 {
		return "no cards found for the user.", nil
	}

	err = s.stolenCardRepository.ReportStolenCards(userID, cardIDs)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: reported_cards.user_id, reported_cards.card_id" {
			return fmt.Sprintf("the report for %s has already been submitted.", userName), nil
		}
		return "", err
	}

	return "all the cards linked to the provided user are now blocked. Contact @support-team for more information.", nil
}

func (s *complianceService) CheckComplianceStatus(userID int64, cardID int64) (bool, string, error) {
	owned, err := s.isCardOwnedByUser(userID, cardID)
	if err != nil {
		return false, "error retrieving user cards", err
	}
	if !owned {
		return false, "the provided card does not belong to the user", nil
	}

	blocked, err := s.stolenCardRepository.IsCardReported(userID, cardID)
	if err != nil {
		return false, "error checking compliance status", err
	}
	if blocked {
		return false, "user is currently blocked due to reported stolen card/s", nil
	}

	return true, "user is compliance", nil
}

func (s *complianceService) isCardOwnedByUser(userID int64, cardID int64) (bool, error) {
	cardIDs, err := s.cardRepository.GetUserCards(userID)
	if err != nil {
		return false, err
	}

	return slices.Contains(cardIDs, cardID), nil
}
