package usecase

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/domain/ports"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/forecast"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
	"github.com/nomenarkt/medicine-tracker/backend/internal/util"
)

// StockChecker handles alerting when stock is near depletion.
type StockChecker struct {
	Airtable ports.AirtableService
	Telegram ports.TelegramService
}

// CheckAndAlertLowStock scans medicines and alerts if 10 days from out-of-stock.
func (s *StockChecker) CheckAndAlertLowStock() error {
	now := time.Now().UTC()
	log.Println("📡 Starting CheckAndAlertLowStock...")

	meds, err := s.Airtable.FetchMedicines()
	if err != nil {
		return fmt.Errorf("fetch medicines failed: %w", err)
	}
	log.Printf("📋 Fetched %d medicines", len(meds))

	entries, err := s.Airtable.FetchStockEntries()
	if err != nil {
		return fmt.Errorf("fetch stock entries failed: %w", err)
	}
	log.Printf("📦 Fetched %d stock entries", len(entries))

	for _, m := range meds {
		stock := stockcalc.CurrentStockAt(m, entries, now)
		if stock <= 0 || m.DailyDose == 0 {
			continue
		}

		forecastDate := stockcalc.OutOfStockDateAt(m, stock, now)
		daysLeft := int(forecastDate.Truncate(24*time.Hour).Sub(now.Truncate(24*time.Hour)).Hours() / 24)

		log.Printf("🔍 %s: stock=%.2f, forecast=%s, daysLeft=%d", m.Name, stock, forecastDate.Format("2006-01-02"), daysLeft)

		if daysLeft <= 10 {
			if m.LastAlertedDate != nil && m.LastAlertedDate.Format("2006-01-02") == now.Format("2006-01-02") {
				log.Printf("ℹ️ Already alerted for %s today, skipping.", m.Name)
				continue
			}

			alert := fmt.Sprintf(
				"*%s* will run out in %d day(s)\\!\nRefill before *%s*\nCurrently: *%.2f* pills left\\.",
				util.EscapeMarkdown(m.Name),
				daysLeft,
				forecastDate.Format("2006-01-02"),
				stock,
			)

			log.Printf("📲 Sending alert for %s", m.Name)
			if err := s.Telegram.SendTelegramMessage(alert); err != nil {
				log.Printf("❌ Telegram send failed: %v", err)
				continue
			}
			log.Println("✅ Telegram message sent")

			if err := s.Airtable.UpdateLastAlertedDate(m.ID, now); err != nil {
				log.Printf("⚠️ Failed to update LastAlertedDate for %s: %v", m.Name, err)
			}
		}
	}

	// 👇 Refill notification logic
	refillsToday := map[string][]domain.StockEntry{}
	for _, entry := range entries {
		if entry.Date.UTC().Format("2006-01-02") == now.Format("2006-01-02") {
			refillsToday[entry.MedicineID] = append(refillsToday[entry.MedicineID], entry)
		}
	}

	for medID, todayEntries := range refillsToday {
		var med *domain.Medicine
		for _, m := range meds {
			if m.ID == medID {
				med = &m
				break
			}
		}
		if med == nil {
			continue
		}

		var lines []string
		for _, e := range todayEntries {
			lines = append(lines,
				fmt.Sprintf("• %d %s on %s",
					e.Quantity,
					util.EscapeMarkdown(e.Unit),
					e.Date.Format("2006-01-02")),
			)
		}

		msg := fmt.Sprintf(
			"*Refill recorded for %s*\\:\n%s",
			util.EscapeMarkdown(med.Name),
			strings.Join(lines, "\n"),
		)

		log.Printf("📲 Notifying refill for %s", med.Name)
		if err := s.Telegram.SendTelegramMessage(msg); err != nil {
			log.Printf("❌ Refill Telegram send failed: %v", err)
		} else {
			log.Printf("✅ Refill message sent for %s", med.Name)
		}
	}
	return nil
}

// OutOfStockService wraps forecast generation logic.
type OutOfStockService struct {
	Airtable ports.StockDataPort
}

// GenerateOutOfStockForecastMessage returns a summary of stock depletion.
func (s OutOfStockService) GenerateOutOfStockForecastMessage() (string, error) {
	meds, err := s.Airtable.FetchMedicines()
	if err != nil {
		return "", fmt.Errorf("fetch medicines failed: %w", err)
	}
	entries, err := s.Airtable.FetchStockEntries()
	if err != nil {
		return "", fmt.Errorf("fetch stock entries failed: %w", err)
	}

	return forecast.GenerateOutOfStockForecastMessage(meds, entries, time.Now().UTC(), s.Airtable), nil
}
