# ⚡ Agentic AI Customer Support & Case Management System

A high-performance, event-driven, production-grade **AI-powered Customer Support & Case Management Platform** built for modern e-commerce marketplaces.

This platform integrates a **Telegram Bot Gateway (Go Service)**, an **Agentic AI Core (Python)** with ReAct Tool Calling & Dual LLM Failover, a **Vector RAG Pipeline (pgvector)** for SOP documents, an automated **Case Escalation & Assignment System**, and a **Real-Time Telemetry & Monitoring CS Dashboard**.

---

## 🌟 Key Features & Architecture Highlights

### 1. 🤖 Multi-Turn Telegram CS Agent (Go & Redis Pub/Sub)
- Event-driven asynchronous architecture connecting Go Service & Python AI Agent via Redis Pub/Sub.
- Telegram **Markdown ParseMode** rendering (`*bold*`, `_italic_`, `` `code` ``) with automatic plain-text retry fallback.
- Session & chat history managed in Redis with rolling window cache (10 turns) and 24-hour TTL (`HISTORY_TTL = 86400`).

### 2. 🔐 Security, Identity & IDOR Protection
- **Backend-Injected Customer ID**: Telegram `chat_id` is deterministically mapped to `tg_{chat_id}` on the server.
- The LLM **never determines or inputs `customer_id`**. The backend automatically injects the verified `customer_id` into all customer-scoped tools (`query_orders`, `get_purchase_total`, `query_cases`, `create_case`). This prevents IDOR vulnerability and data leakage between customers.
- **Determinisitc Guardrails**: Rule-based Regex input filtering (SQLi, Prompt Injection, Jailbreak attempts, PII protection, message length limits) and output sanitization.

### 3. 🧠 Agentic AI Core & Tools (Dual LLM Failover)
- **Dual-Provider Failover**: Primary LLM (SK Guts / OpenAI / Gemini) with automatic seamless failover to Fallback LLM (OpenRouter).
- **Comprehensive Tool Definitions**:
  - `query_orders`: Smart order lookup & tracking per customer.
  - `get_purchase_total`: Calculate annual/total spending amounts.
  - `query_cases`: Real-time status checking of complaint tickets (`open`, `in_progress`, `resolved`, `closed`).
  - `create_case`: Auto-escalates complex complaints to human staff (`payment`, `shipping`, `IT` divisions) returning UUID Case IDs.
  - `search_sop`: Retrieval-Augmented Generation (RAG) vector similarity search on company SOPs.

### 4. 📚 Dynamic SOP Knowledge Base & RAG Pipeline (Word & PDF Upload)
- Admin can upload SOP documents directly from Web Dashboard in **Word (`.docx`)**, **PDF (`.pdf`)**, or **Text (`.txt`)** format.
- Instant client-side text extraction & automatic vector embedding generation (`pgvector` cosine similarity) **without restarting Docker containers**.
- Full CRUD (Create, Read, Update/Edit, Delete) for SOP Knowledge Base.

### 5. 👥 User & Division/Role Management
- Centralized staff management on Web Dashboard for Divisions: **`payment`**, **`shipping`**, and **`IT`**.
- Automatic matching and assignment (`assign_to`) of created tickets to staff members of the appropriate division.

### 6. 📊 Real-Time Telemetry & Monitoring CS Dashboard
- Single-Page Web Dashboard at `http://localhost:8080/dashboard`.
- **Top Global Stat Cards**: Average Latency (ms), Total AI Requests, Token Usage Breakdown (Prompt & Completion Tokens), System Errors.
- **AI Tool Calls Summary**: Breakdown of tool call counts and execution times.
- **Live Trace Telemetry Table**: Real-time per-request trace logs displaying `Trace ID`, `Chat ID`, `Latency (ms)`, `Tokens (Prompt/Completion)`, `Tool Calls Executed`, `Errors`, and `Timestamp`.

---

## 🛠️ Architecture & Technology Stack

```
Customer ──► Telegram Bot API ──► Go Gateway (Chi HTTP Router)
                                        │ (Redis Pub/Sub)
                                        ▼
                                 Python AI Agent Core
                                 [Guardrail Layer]
                                        │
           ┌────────────────────────────┼────────────────────────────┐
           ▼                            ▼                            ▼
      DB Order                    Doc SOP (RAG)             Case Management &
  (Parameterized SQL)        (pgvector Embeddings)       Admin CS Web Dashboard
```

- **Gateway & Backend**: Go 1.22+ (Chi Router, `pgxpool`, `go-redis`, `telegram-bot-api`)
- **AI Core & RAG**: Python 3.12+ (OpenAI SDK, `pgvector`, `psycopg2`, Mammoth.js, PDF.js)
- **Database**: PostgreSQL 16 + `pgvector` extension
- **Queue & Multi-turn Cache**: Redis 7 (Pub/Sub & List with 24h TTL)

---

## 🗄️ Database Schemas

### 1. `Doc` (SOP Knowledge Base)
- `uuid` (UUID, Primary Key)
- `name` (VARCHAR, e.g. "SOP Pengembalian Barang")
- `doc` (TEXT, Content of SOP)
- `embedding` (VECTOR 1536)
- `created_at`, `updated_at` (TIMESTAMP)

### 2. `Case` (Complaint Tickets)
- `uuid` (UUID, Primary Key)
- `name` (VARCHAR, Case Title)
- `datetime` (TIMESTAMP, Creation Time)
- `criticality` (VARCHAR: `low`, `medium`, `high`, `critical`)
- `description` (TEXT, Customer Complaint & ID)
- `assign_to` (UUID / TEXT, Assigned Staff & Division)
- `status` (VARCHAR: `open`, `in_progress`, `resolved`, `closed`)
- `resolution` (TEXT, Resolution notes)

### 3. `User` (Internal Staff)
- `uuid` (UUID, Primary Key)
- `name` (VARCHAR, Staff Name)
- `division` (VARCHAR: `payment`, `shipping`, `IT`)

### 4. `Order` (Transactions Data)
- `uuid` (UUID, Primary Key)
- `customer_id` (VARCHAR, Telegram Customer ID)
- `product_name` (VARCHAR)
- `amount` (BIGINT)
- `status` (VARCHAR: `shipped`, `delivered`, `processing`, `cancelled`)
- `order_date` (TIMESTAMP)
- `tracking_id` (VARCHAR)

---

## 🚀 Quick Start (Production Setup with Docker)

### 1. Environment Configuration
Copy `.env.example` to `.env` and configure your API Keys:
```bash
cp .env.example .env
```
Example `.env`:
```env
DATABASE_URL=postgres://user:password@postgres:5432/agenticsdk?sslmode=disable
REDIS_URL=redis://redis:6379
SERVER_PORT=8080
GO_SERVICE_URL=http://go-service:8080

TELEGRAM_BOT_TOKEN=your-telegram-bot-token

# Primary LLM Provider
LLM_API_KEY=sk-your-primary-llm-key
LLM_BASE_URL=https://openrouter.ai/api/v1
LLM_MODEL=google/gemini-2.5-flash:free

# Fallback LLM Provider
FALLBACK_LLM_API_KEY=sk-your-fallback-key
FALLBACK_LLM_BASE_URL=https://openrouter.ai/api/v1
FALLBACK_LLM_MODEL=google/gemini-2.5-flash:free

# Vector Embeddings Provider
OPENAI_EMBEDDING_API_KEY=sk-your-embedding-key
OPENAI_EMBEDDING_BASE_URL=https://openrouter.ai/api/v1
```

### 2. Run the Entire Stack with Docker Compose
```bash
docker compose up -d --build
```

### 3. Run Automated Database Migration & Seeding
```bash
bash scripts/setup.sh
```

---

## 💻 Admin Dashboard & System Access

- **Web Dashboard**: Open [http://localhost:8080/dashboard](http://localhost:8080/dashboard) (or `http://YOUR_VPS_IP:8080/dashboard`)
  - **📋 Case Management**: View complaint tickets, change status (`open`, `in_progress`, `resolved`, `closed`), and view assigned division/staff.
  - **👥 Kelola Staf & Divisi**: Add/remove staff users and assign roles/divisions (`payment`, `shipping`, `IT`).
  - **📊 Monitoring & Traces**: View real-time AI Agent traces, token usage, latency, tool calls, and error logs.
  - **📚 Knowledge Base & SOP (RAG)**: Upload Word (`.docx`), PDF (`.pdf`), or Text (`.txt`) SOPs, edit existing SOPs, and trigger vector re-indexing.
  - **📦 Order Data Seeder**: Inject mock test orders to test Telegram bot queries.

---

## 🧪 Integration Testing

Run the automated E2E integration test suite:
```bash
python scripts/test_e2e.py
```
