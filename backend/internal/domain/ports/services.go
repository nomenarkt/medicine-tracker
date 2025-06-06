package ports

import (
	"context"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
)

type AirtableService interface {
	FetchMedicines() ([]domain.Medicine, error)
	FetchStockEntries() ([]domain.StockEntry, error)
	UpdateLastAlertedDate(medicineID string, date time.Time) error
}

type TelegramService interface {
	SendTelegramMessage(text string) error
	PollForCommands(ctx context.Context, fetch func() ([]domain.Medicine, []domain.StockEntry, error))
}

type StockDataPort interface {
	FetchMedicines() ([]domain.Medicine, error)
	FetchStockEntries() ([]domain.StockEntry, error)
	CreateStockEntry(domain.StockEntry) error
	UpdateForecastDate(medicineID string, forecastDate, updatedAt time.Time) error
}
