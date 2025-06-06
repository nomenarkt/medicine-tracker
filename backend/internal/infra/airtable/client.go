package airtable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
)

type Client struct{}

func NewClient() *Client {
	_ = godotenv.Load()
	return &Client{}
}

type airtableRecord[T any] struct {
	ID     string `json:"id"`
	Fields T      `json:"fields"`
}

type airtableResponse[T any] struct {
	Records []airtableRecord[T] `json:"records"`
}

func (c *Client) FetchMedicines() ([]domain.Medicine, error) {
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s",
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_MEDICINES_TABLE"))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Println("airtable response close error:", cerr)
		}
	}()

	body, _ := io.ReadAll(res.Body)

	var errCheck map[string]interface{}
	if json.Unmarshal(body, &errCheck) == nil {
		if errVal, exists := errCheck["error"]; exists {
			return nil, fmt.Errorf("airtable error: %v", errVal)
		}
	}

	var response airtableResponse[domain.Medicine]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	var meds []domain.Medicine
	for _, rec := range response.Records {
		m := rec.Fields
		m.ID = rec.ID
		meds = append(meds, m)
	}
	return meds, nil
}

func (c *Client) FetchStockEntries() ([]domain.StockEntry, error) {
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s",
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_ENTRIES_TABLE"))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Println("airtable response close error:", cerr)
		}
	}()

	body, _ := io.ReadAll(res.Body)

	var errCheck map[string]interface{}
	if json.Unmarshal(body, &errCheck) == nil {
		if errVal, exists := errCheck["error"]; exists {
			return nil, fmt.Errorf("airtable error: %v", errVal)
		}
	}

	var response airtableResponse[domain.StockEntry]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	var entries []domain.StockEntry
	for _, rec := range response.Records {
		e := rec.Fields
		e.ID = rec.ID
		entries = append(entries, e)
	}
	return entries, nil
}

func (c *Client) CreateStockEntry(entry domain.StockEntry) error {
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s",
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_ENTRIES_TABLE"))

	payload := map[string]any{
		"fields": map[string]any{
			"medicine_id": []string{entry.MedicineID},
			"quantity":    entry.Quantity,
			"unit":        entry.Unit,
			"date":        entry.Date.Format("2006-01-02"),
		},
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Println("airtable response close error:", cerr)
		}
	}()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("airtable error: %s", string(b))
	}

	return nil
}

func (c *Client) UpdateForecastDate(medicineID string, forecastDate, updatedAt time.Time) error {
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s",
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_MEDICINES_TABLE"),
		medicineID)

	payload := map[string]any{
		"fields": map[string]any{
			"forecast_out_of_stock_date": forecastDate.Format("2006-01-02"), // ✅ always date-only
			"forecast_last_updated":      updatedAt.Format("2006-01-02"),
		},
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Println("airtable response close error:", cerr)
		}
	}()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("airtable error: %s", string(b))
	}

	return nil
}

func (c *Client) UpdateLastAlertedDate(medicineID string, date time.Time) error {
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s",
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_MEDICINES_TABLE"),
		medicineID)

	payload := map[string]any{
		"fields": map[string]any{
			"last_alerted_date": date.Format("2006-01-02"),
		},
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Println("airtable response close error:", cerr)
		}
	}()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("airtable error: %s", string(b))
	}

	return nil
}
