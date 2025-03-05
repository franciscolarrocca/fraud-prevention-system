package handler

import (
	"flarrocca/compliant-service/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ComplianceHandler struct {
	complianceService service.ComplianceService
}

func NewUserHandler(complianceService service.ComplianceService) *ComplianceHandler {
	return &ComplianceHandler{complianceService: complianceService}
}

func (h *ComplianceHandler) ReportStolenCards(c *fiber.Ctx) error {
	userName := c.FormValue("user_name")
	secretCode := c.FormValue("secret_code")
	if userName == "" || secretCode == "" {
		return c.Status(http.StatusBadRequest).SendString("user name and secret code are required")
	}

	message, err := h.complianceService.ReportStolenCards(userName, secretCode)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString(message)
}

func (h *ComplianceHandler) CheckComplianceStatus(c *fiber.Ctx) error {
	paramUserID := c.Query("user_id")
	if paramUserID == "" {
		return c.Status(http.StatusBadRequest).SendString("user id is required")
	}

	paramCardID := c.Query("card_id")
	if paramCardID == "" {
		return c.Status(http.StatusBadRequest).SendString("card id is required")
	}

	userID, err := strconv.Atoi(paramUserID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"complaiance": false,
			"message":     fmt.Sprintf("invalid data type for user ID: %s", err),
		})
	}

	cardID, err := strconv.Atoi(paramCardID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"complaiance": false,
			"message":     fmt.Sprintf("invalid data type for card ID: %s", err),
		})
	}

	isCompliance, message, err := h.complianceService.CheckComplianceStatus(int64(userID), int64(cardID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"complaiance": isCompliance,
			"message":     fmt.Sprintf("error checking user status: %s", message),
		})
	}

	return c.JSON(fiber.Map{
		"complaiance": isCompliance,
		"message":     message,
	})
}
