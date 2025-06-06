
# 💊 Medicine Tracker

A reliable, real-time medicine stock tracker with:

- ✅ Dynamic pill stock forecasting  
- 📆 Out-of-stock date prediction  
- 🔁 Refill handling (idempotent & date-bound)  
- 🤖 Telegram alerts  
- 📊 Airtable-based data storage  

---

## 🛠️ Stack

| Layer      | Tech                        |
|------------|-----------------------------|
| Backend    | Go (Fiber)                  |
| Data Store | Airtable (REST API)         |
| Alerts     | Telegram Bot API            |
| Deployment | Render (free-tier Docker)   |

---

## 🧱 Why This Stack?

Our stack is optimized for simplicity, cost-efficiency, and real-time alerting.

| Layer      | Tech             | Why? |
|------------|------------------|------|
| **Backend** | Go (Fiber)        | Fast, minimal memory footprint, perfect for APIs. Fiber is expressive but fast like Node.js. |
| **Data Store** | Airtable REST API | Great for prototyping with a spreadsheet-like UI and flexible schema. |
| **Alerts** | Telegram Bot API | Easy setup, excellent UX for push notifications, and no cost. |
| **Deploy** | Render.com        | Free-tier hosting with Docker support, zero-config, and smooth scaling. |

This stack enables lean infrastructure with real-time automation and friendly UX — ideal for side projects, MVPs, and internal tools.

---

## 🔧 Configuration

Copy `backend/.env.template` to `backend/.env` and fill in your credentials:

```env
AIRTABLE_BASE_ID=
AIRTABLE_MEDICINES_TABLE=
AIRTABLE_ENTRIES_TABLE=
AIRTABLE_TOKEN=
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
ENABLE_ENTRY_POST=false
ENABLE_ALERT_TICKER=false
```

---

## 📦 Alerting & Notifications

### ⏰ Out-of-Stock Alerts

The system checks every medicine daily:

- Sends a **Telegram alert** when **10 days or less** remain before depletion.
- Skips the alert if already sent today (tracked via `LastAlertedDate` field in Airtable).

**Alert Format:**

```
*<Medicine>* will run out in <X> day(s)\!
Refill before *YYYY-MM-DD*
Currently: *N.NN* pills left\.
```

### ♻️ Refill Detection

If a stock entry was created **today**, a Telegram **refill notification** is sent:

**Refill Format:**

```
*Refill recorded for <Medicine>*:
• 2 box on 2025-06-01
```

- Only today's entries are considered.
- Refill messages are **separate** from depletion alerts.

### ⛔ Alert Suppression Rules

- Only one alert per medicine per day.
- `LastAlertedDate` prevents duplicates.
- Refills are processed only if the `date == today`.

---

## 🧪 Testing

Run test suite with:

```bash
make test
```

Includes:

- Unit tests for stock calculation
- Table-driven tests for alert triggering & refill notifications

---

## 🔍 Example: Telegram `/stock`

The `/stock` command uses `CurrentStockAt()` and shows dynamic forecast:

```
*Out-of-Stock Forecast*

```text
MedA                  → 2025-06-10 (12.00 left)
MedB                  → 2025-06-15 (28.00 left)
```
```

---

## 📁 Key Modules

```
internal/
├── usecase/alert.go        # All alert + refill logic
├── logic/stockcalc/        # Computes current stock
├── infra/telegram/         # Telegram /stock handler
├── infra/airtable/         # API integration layer
├── domain/models.go        # Medicine struct + forecast fields
```

---

## 📎 Notes

- All Telegram messages are MarkdownV2-safe.
- Refill notifications and alerts are **idempotent**.
- `LastAlertedDate` ensures alerts are not duplicated.

## License

Released under the [MIT License](LICENSE).

---
