package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"io"
	"net/url"
	"sort"
	"time"

	"github.com/joho/godotenv"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
	"github.com/nomenarkt/medicine-tracker/backend/internal/util"
)

type Client struct {
	Token  string
	ChatID string
}

func NewClient() *Client {
	_ = godotenv.Load()
	return &Client{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID: os.Getenv("TELEGRAM_CHAT_ID"),
	}
}

func (c *Client) SendTelegramMessage(msg string) error {
	log.Printf("📨 Sending Telegram: %s", msg)

	escaped := msg
	if !strings.Contains(msg, "```") {
		escaped = util.EscapeMarkdown(msg)
	}

	payload := map[string]string{
		"chat_id":    c.ChatID,
		"text":       escaped,
		"parse_mode": "MarkdownV2",
	}

	body, _ := json.Marshal(payload)

	res, err := http.Post(
		"https://api.telegram.org/bot"+c.Token+"/sendMessage",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fmt.Errorf("telegram error status: %d", res.StatusCode)
	}

	return nil
}

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

type GetUpdatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

func (c *Client) PollForCommands(ctx context.Context, fetchData func() ([]domain.Medicine, []domain.StockEntry, error)) {
	var lastUpdateID int

	log.Println("📨 Telegram polling started...")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(2 * time.Second)
		}

		apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?timeout=10&offset=%d", c.Token, lastUpdateID+1)
		resp, err := http.Get(apiURL)
		if err != nil {
			log.Println("Telegram polling error:", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var updates GetUpdatesResponse
		if err := json.Unmarshal(body, &updates); err != nil {
			log.Println("Failed to decode Telegram updates:", err)
			continue
		}
		if !updates.OK {
			log.Printf("Telegram API error status %d: %s", resp.StatusCode, string(body))
			continue
		}

		for _, update := range updates.Result {
			lastUpdateID = update.UpdateID
			if update.Message.Text == "/stock" {
				go c.handleStockCommand(update.Message.Chat.ID, fetchData)
			}
		}
	}
}

func (c *Client) handleStockCommand(chatID int64, fetchData func() ([]domain.Medicine, []domain.StockEntry, error)) {
	meds, entries, err := fetchData()
	if err != nil {
		_ = c.sendTo(chatID, "\u26a0\ufe0f Failed to fetch stock data.")
		return
	}

	now := time.Now()
	type Row struct {
		Name  string
		Date  time.Time
		Pills float64
	}
	var rows []Row
	for _, m := range meds {
		stock := stockcalc.CurrentStockAt(m, entries, now)
		if stock <= 0 || m.DailyDose == 0 {
			continue
		}
		date := stockcalc.OutOfStockDateAt(m, stock, now)
		rows = append(rows, Row{m.Name, date, stock})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Date.Before(rows[j].Date)
	})

	var lines []string
	for _, r := range rows {
		lines = append(lines, fmt.Sprintf("%-22s → %s (%.2f left)", r.Name, r.Date.Format("2006-01-02"), r.Pills))
	}

	msg := "*Out-of-Stock Forecast*\n\n```text\n" + strings.Join(lines, "\n") + "\n```"
	_ = c.sendTo(chatID, msg)
}

func (c *Client) sendTo(chatID int64, msg string) error {
	payload := url.Values{}
	payload.Set("chat_id", fmt.Sprintf("%d", chatID))
	payload.Set("text", msg)
	payload.Set("parse_mode", "Markdown")

	_, err := http.PostForm(
		"https://api.telegram.org/bot"+c.Token+"/sendMessage",
		payload,
	)
	return err
}
