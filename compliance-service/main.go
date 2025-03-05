package main

import (
	"database/sql"
	"flarrocca/compliant-service/handler"
	"flarrocca/compliant-service/repository"
	"flarrocca/compliant-service/service"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./database/compliance.db")
	if err != nil {
		log.Fatal(err)
	}

	initSQL, err := os.ReadFile("./database/init.sql")
	if err != nil {
		log.Fatal("error reading init.sql:", err)
	}

	_, err = db.Exec(string(initSQL))
	if err != nil {
		log.Fatal("error executing init.sql:", err)
	}

	return db
}

func main() {
	db := initDB()

	userRepository := repository.NewUserRepository(db)
	cardRepository := repository.NewCardRepository(db)
	stolenCardRepository := repository.NewStolenCardRepository(db)
	complianceService := service.NewComplianceService(userRepository, cardRepository, stolenCardRepository)
	complianceHandler := handler.NewUserHandler(complianceService)

	tmplEngine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{Views: setVueCompatibleDelimiters(tmplEngine)})

	app.Static("/static", "./views/static")
	app.Get("/report", func(c *fiber.Ctx) error {
		return c.Render("report", fiber.Map{})
	})

	app.Post("/report_cards", complianceHandler.ReportStolenCards)
	app.Get("/check_user", complianceHandler.CheckComplianceStatus)

	log.Fatal(app.Listen(":8080"))
}

// setVueCompatibleDelimiters avoid conflicts with Vue.js {{ }}
func setVueCompatibleDelimiters(tmplEngine *html.Engine) *html.Engine {
	tmplEngine.Delims("<<", ">>")
	return tmplEngine
}
