package main

import (
	"flarrocca/payment-service/handler"
	"flarrocca/payment-service/repository"
	"flarrocca/payment-service/service"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	complianceRepository := repository.NewComplianceRepository()

	paymentProcessorService := service.NewPaymentProcessorService(complianceRepository)
	paymentProcessorHandler := handler.NewPaymentProcessorHandler(paymentProcessorService)

	app.Post("/process_payment", paymentProcessorHandler.ProcessPayment)

	log.Fatal(app.Listen(":8081"))
}
