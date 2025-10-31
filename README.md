# Bank Statement Processor

## Features

- **Streaming CSV Processing** - Processes large files line-by-line without loading into memory
- **Asynchronous Event Processing** - Failed transactions are published to an in-memory event bus
- **Worker Pool with Retry Logic** - Reconciliation consumer with exponential backoff
- **Concurrent Safe** - All operations are thread-safe (passes race detector)
- **Clean Architecture** - Clear separation of concerns with dependency injection
- **Graceful Shutdown** - Finishes in-flight work before stopping
- **Structured Logging** - JSON logs for observability

## Architecture Overview
```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────────────────────────────────┐
│         HTTP Handler Layer              │
│  (statement, balance, issues, health)   │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│          Use Case Layer                 │
│  (business logic orchestration)         │
└──────┬──────────────────┬───────────────┘
       │                  │
       ▼                  ▼
┌──────────────┐   ┌──────────────┐
│  Repository  │   │  Event Bus   │
│  (in-memory) │   │  (channels)  │
└──────────────┘   └──────┬───────┘
                          │
                          ▼
                   ┌──────────────┐
                   │  Reconcile   │
                   │   Consumer   │
                   └──────────────┘
```
## Getting Started

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile)

### Installation
```bash
# Clone the repository
git clone https://github.com/yourusername/bank-statement-processor.git
cd bank-statement-processor

# Download dependencies
make deps
# or
go mod download
```

### Running the Application
```bash
# Using Makefile
make run

# Or directly with Go
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

### Running Tests
```bash
# Run tests with race detector
make test
```

## API Endpoints

### 1. Upload Statement

Upload a CSV file for processing. Sample CSV file -> 'test_statement_csv_100.csv'

**Request:**
```http
POST /statements
Content-Type: multipart/form-data
```

**Response:**
```json
{
  "upload_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "CSV upload accepted and processing started"
}
```

**Status Codes:**
- `202 Accepted` - Upload accepted and processing started
- `400 Bad Request` - Invalid file or missing parameters
- `413 Request Entity Too Large` - File exceeds 100MB limit
- `500 Internal Server Error` - Server error

---

### 2. Get Balance

Retrieve the balance for an uploaded statement.

**Request:**
```http
GET /balance?upload_id={upload_id}
```

**Response (Processing):**
```json
{
  "upload_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "balance": null,
  "message": "CSV is still being processed"
}
```

**Response (Completed):**
```json
{
  "upload_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "balance": 16542000
}
```

**Balance Calculation:**
- Only `SUCCESS` transactions are included
- `FAILED` and `PENDING` transactions are excluded

**Status Codes:**
- `200 OK` - Balance retrieved successfully
- `400 Bad Request` - Missing upload_id
- `404 Not Found` - Upload not found

---

### 3. Get Issues

List problematic transactions (FAILED and PENDING).

**Request:**
```http
GET /transactions/issues?upload_id={upload_id}&status={status}&page={page}&page_size={page_size}
```

**Query Parameters:**
- `upload_id` (required): Upload identifier
- `status` (optional): Filter by `FAILED` or `PENDING`
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)
- `min_amount` (optional): Minimum transaction amount
- `max_amount` (optional): Maximum transaction amount
- `from_date` (optional): Start timestamp (Unix seconds)
- `to_date` (optional): End timestamp (Unix seconds)

**Response:**
```json
{
  "upload_id": "550e8400-e29b-41d4-a716-446655440000",
  "transactions": [
    {
      "id": "tx-123",
      "timestamp": 1674509012,
      "counterparty": "ELECTRIC COMPANY",
      "type": "DEBIT",
      "amount": 450000,
      "status": "FAILED",
      "description": "utility bill"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 8,
    "total_pages": 1
  }
}
```

**Status Codes:**
- `200 OK` - Issues retrieved successfully
- `400 Bad Request` - Invalid parameters

---

### 4. Health Check

Check if the service is healthy.

**Request:**
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1674507883
}
```

**Status Codes:**
- `200 OK` - Service is healthy

## Usage Examples

### Upload a CSV File
```bash
curl -X POST \
  -F "file=@examples/sample_statement.csv" \
  http://localhost:8080/statements
```

**Response:**
```json
{
  "upload_id": "abc123",
  "message": "CSV upload accepted and processing started"
}
```

---

### Check Balance (While Processing)
```bash
curl "http://localhost:8080/balance?upload_id=abc123"
```

**Response:**
```json
{
  "upload_id": "abc123",
  "status": "processing",
  "balance": null,
  "message": "CSV is still being processed"
}
```

---

### Check Balance (After Processing)
```bash
curl "http://localhost:8080/balance?upload_id=abc123"
```

**Response:**
```json
{
  "upload_id": "abc123",
  "status": "completed",
  "balance": 16542000
}
```

---

### Get All Issues
```bash
curl "http://localhost:8080/transactions/issues?upload_id=abc123"
```

---

### Get Only FAILED Transactions
```bash
curl "http://localhost:8080/transactions/issues?upload_id=abc123&status=FAILED"
```

---

### Get Issues with Pagination
```bash
curl "http://localhost:8080/transactions/issues?upload_id=abc123&page=1&page_size=10"
```

---

### Filter by Amount
```bash
# Get failed transactions over 1,000,000
curl "http://localhost:8080/transactions/issues?upload_id=abc123&status=FAILED&min_amount=1000000"
```

---

### Filter by Date Range
```bash
curl "http://localhost:8080/transactions/issues?upload_id=abc123&from_date=1674507883&to_date=1674600000"
```

---

### Combine Multiple Filters
```bash
curl "http://localhost:8080/transactions/issues?upload_id=abc123&status=FAILED&min_amount=500000&page=1&page_size=20"
```

---

### Health Check
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1674507883
}
```

## Design Decisions & Tradeoffs

### 1. Buffered Channels for Event Bus
**Decision:** Use buffered channels (size: 100) with blocking behavior

**Pros:**
- Absorbs bursts of events
- Decouples publisher from consumer
- No events lost (blocks if buffer full)

**Cons:**
- CSV processing may slow if consumers are very slow
- Memory overhead for buffer

**Alternative:** Could drop events with warning

---

### 2. Balance Calculation
**Decision:** Return calculated balance

**Pros:**
- O(1) time complexity per query
- No recalculates on every request

**Cons:**
- Hard to maintain consistency when there is additional requirements related to balance

**Alternative:** Calculate on-demand by iterating transactions

---

### 3. Reconciliation Consumer
**Decision:** Simulate reconciliation with logging and retry logic

**Pros:**
- Demonstrates event-driven patterns
- Shows retry/backoff implementation
- Shows idempotency handling

**Cons:**
- Doesn't actually update transaction status
- Balance doesn't change after reconciliation

---

### 4. Graceful Shutdown
**Decision:** Use context cancellation and wait groups

**Pros:**
- Respects in-flight work
- Clean shutdown of goroutines
- Prevents data loss

**Implementation:**
- HTTP server shutdown with 30s timeout
- Background workers respect context cancellation
- Event bus closes all subscriber channels

---

## Event Processing Flow
```
1. CSV Upload
   ↓
2. Parse CSV line-by-line (streaming)
   ↓
3. For each transaction:
   - Save to repository
   - If FAILED → Publish event to bus
   ↓
4. Event Bus
   ↓
5. Worker Pool (3 workers)
   ↓
6. Reconciliation Consumer
   - Retry with backoff (1s, 2s, 4s)
   - Idempotent processing
   - Structured logging
```

## Observability

### Structured Logging

All components use structured logging:
```
2024/01/20 10:15:23 {"time":"2025-10-31 18:19:51.900850891 +0700 WIB m=+0.000542605","level":"info","message":"starting http server at:8080"}
2024/01/20 10:15:30 {"time":"2025-10-31 18:19:51.90086106 +0700 WIB m=+0.000552774","level":"info","message":"starting reconciliation consumer with 3 workers"}
2024/01/20 10:15:35 {"time":"2025-10-31 18:19:51.901114424 +0700 WIB m=+0.000806138","level":"info","message":"new subscriber added (total: 1)"}
2024/01/20 10:15:36 {"time":"2025-10-31 18:19:51.901185657 +0700 WIB m=+0.000877371","level":"info","message":"reconciliation worker 2 started"}
```

### Health Endpoint

The `/health` endpoint can be used for:
- Load balancer health checks
- Monitoring systems

## Testing

### Unit Tests

Tests cover business logic:
- CSV parsing validation
- Balance calculation rules
- Issue filtering logic

### Integration Tests

End-to-end workflow tests:
- Upload → Process → Query balance
- Upload → Query issues
- Verify reconciliation

### Race Detection

All tests pass with race detector:
```bash
make test
```