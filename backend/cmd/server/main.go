package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/nomenarkt/medicine-tracker/backend/internal/background"
	"github.com/nomenarkt/medicine-tracker/backend/internal/di"
	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/telegram"
	"github.com/nomenarkt/medicine-tracker/backend/internal/server"
)

func main() {
	_ = godotenv.Load()
	app := fiber.New()

	// ⛓️ Resolve dependencies via central initializer
	deps := di.Init()

	// ✅ Setup all HTTP routes with DI
	server.SetupRoutes(app, deps.StockChecker, deps.ForecastSvc, deps.Airtable, deps.Telegram)

	// 🔄 Start background stock check (daily) if enabled
	if os.Getenv("ENABLE_ALERT_TICKER") == "true" {
		background.StartStockAlertTicker(telegram.HandleOutOfStockCommand)
	}

	// 🧭 Start Telegram bot polling for `/stock` commands
	go deps.Telegram.PollForCommands(func() ([]domain.Medicine, []domain.StockEntry, error) {
		meds, err := deps.Airtable.FetchMedicines()
		if err != nil {
			return nil, nil, err
		}
		entries, err := deps.Airtable.FetchStockEntries()
		if err != nil {
			return nil, nil, err
		}
		return meds, entries, nil
	})

	// 🚀 Run server
	log.Fatal(app.Listen(":8787"))
}
