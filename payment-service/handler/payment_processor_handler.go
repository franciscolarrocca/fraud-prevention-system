package handler

import (
	"flarrocca/payment-service/service"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type PaymentProcessorHandler struct {
	paymentService service.PaymentProcessorService
}

func NewPaymentProcessorHandler(complianceService service.PaymentProcessorService) *PaymentProcessorHandler {
	return &PaymentProcessorHandler{paymentService: complianceService}
}

func (p *PaymentProcessorHandler) ProcessPayment(c *fiber.Ctx) error {
	var req struct {
		UserID int64   `json:"user_id"`
		CardID int64   `json:"card_id"`
		Amount float64 `json:"amount"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "invalid request payload"})
	}

	if req.UserID == 0 || req.CardID == 0 || req.Amount <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "user id, card id and valid amount are required"})
	}

	message, err := p.paymentService.ProcessPayment(req.UserID, req.CardID, req.Amount)
	if err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": message})
}
