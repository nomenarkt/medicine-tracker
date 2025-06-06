package stockcalc_test

import (
	"testing"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
)

func mustDate(s string) domain.FlexibleDate {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return domain.NewFlexibleDate(t)
}

func TestCurrentStockAt_WithRefillOnToday(t *testing.T) {
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)

	med := domain.Medicine{
		ID:           "med123",
		Name:         "Paracetamol",
		StartDate:    domain.NewFlexibleDate(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
		InitialStock: 10,
		DailyDose:    1,
		UnitPerBox:   10,
	}

	entries := []domain.StockEntry{
		{
			MedicineID: "med123",
			Quantity:   1,
			Unit:       "box",
			Date:       domain.NewFlexibleDate(now),
		},
	}

	got := stockcalc.CurrentStockAt(med, entries, now)
	want := float64(10 - 3 + 10) // used 3 doses (June 2,3,4) + 1 box refill

	if got != want {
		t.Errorf("Expected stock %.2f, got %.2f", want, got)
	}
}

func TestCurrentStockAt_WithMultipleEntryDates(t *testing.T) {
	med := domain.Medicine{
		ID:           "med1",
		Name:         "TestMed",
		UnitPerBox:   10,
		DailyDose:    1.0,
		StartDate:    domain.NewFlexibleDate(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
		InitialStock: 5,
	}

	today, _ := time.Parse("2006-01-02", "2025-06-04")

	entries := []domain.StockEntry{
		{MedicineID: "med1", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(today)},                    // +10
		{MedicineID: "med1", Quantity: 5, Unit: "pill", Date: domain.NewFlexibleDate(today)},                   // +5
		{MedicineID: "med1", Quantity: 5, Unit: "pill", Date: domain.NewFlexibleDate(today.AddDate(0, 0, 1))},  // future: ignored
		{MedicineID: "med1", Quantity: 5, Unit: "pill", Date: domain.NewFlexibleDate(today.AddDate(0, 0, -1))}, // past: ignored
	}

	stock := stockcalc.CurrentStockAt(med, entries, today)

	expected := 5.0 - 3.0 + 15.0 // initial - 3 days used + today's refill
	if stock != expected {
		t.Errorf("Expected %.2f, got %.2f", expected, stock)
	}
}

func TestOutOfStockDateAt(t *testing.T) {
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)

	med := domain.Medicine{
		ID:        "med123",
		DailyDose: 2,
	}

	stock := 10.0
	got := stockcalc.OutOfStockDateAt(med, stock, now)
	want := now.AddDate(0, 0, 5)

	if !got.Equal(want) {
		t.Errorf("Expected out-of-stock date %v, got %v", want, got)
	}
}

func TestCurrentStockAt_WithRFC3339StartDate(t *testing.T) {
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)

	med := domain.Medicine{
		ID:           "medRFC",
		Name:         "RFCMed",
		StartDate:    domain.NewFlexibleDate(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
		InitialStock: 10,
		DailyDose:    1,
		UnitPerBox:   10,
	}

	entries := []domain.StockEntry{
		{
			MedicineID: "medRFC",
			Quantity:   1,
			Unit:       "box",
			Date:       domain.NewFlexibleDate(now),
		},
	}

	got := stockcalc.CurrentStockAt(med, entries, now)
	want := float64(10 - 3 + 10) // used 3 doses + 1 box refill

	if got != want {
		t.Errorf("Expected stock %.2f, got %.2f", want, got)
	}
}

func TestCurrentStockAt_EntryDateRFC3339Match(t *testing.T) {
	start := "2025-06-01"
	now := time.Date(2025, 6, 4, 12, 0, 0, 0, time.UTC)

	med := domain.Medicine{
		ID:           "med2",
		Name:         "AdvancedMed",
		StartDate:    mustDate(start),
		InitialStock: 5,
		DailyDose:    1,
		UnitPerBox:   10,
	}

	rfcDate := time.Date(2025, 6, 4, 12, 0, 0, 0, time.UTC)

	entries := []domain.StockEntry{
		{MedicineID: "med2", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(rfcDate)},
	}

	got := stockcalc.CurrentStockAt(med, entries, now)
	want := 5.0 - 3.0 + 10.0

	if got != want {
		t.Errorf("Expected %.2f, got %.2f", want, got)
	}
}
