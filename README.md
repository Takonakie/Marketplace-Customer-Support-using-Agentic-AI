# Agentic AI Customer Support & Case Management System

A high-performance, event-driven **AI-powered Customer Support & Case Management platform** built for modern marketplaces. 

This project integrates a **Telegram Bot** (Go) for customer interactions, a **Python AI Agent** powered by LLMs (OpenAI / OpenRouter / Gemini) with Tool Calling, a **RAG Pipeline** for searching SOP documents, an automated **Case Management System**, and a **Real-Time Admin CS & Monitoring Dashboard**.

---

## 🌟 Key Features

1. **Telegram Customer Support Bot**: Multi-turn conversation handling via Go Service & Redis Pub/Sub queue.
2. **AI Agent with Tool Calling**:
   - `query_orders`: Smart order lookup per customer.
   - `get_purchase_total`: Calculate annual spending totals.
   - `search_sop`: Retrieval-Augmented Generation (RAG) for finding company SOPs.
   - `create_case`: Escalates unresolvable issues directly to internal teams (Payment, Shipping, IT).
3. **RAG Pipeline (pgvector)**: Vector similarity search for Indonesian slang, informal queries, and official document lookup.
4. **Real-time CS Admin & Observability Dashboard**: Single-page web dashboard at `http://localhost:8080/dashboard` for managing tickets, updating case statuses, and tracking AI metrics.
5. **Guardrails & Safety**: Input length validation, prompt injection protection, and output data leakage masking.

---

## 🛠️ Architecture & Tech Stack

```
Customer → Telegram Bot (Go) ──(Redis Pub/Sub)──► AI Agent (Python)
                                                     │
                                            ┌────────┼────────┐
                                            ▼        ▼        ▼
                                       DB Order   Doc SOP    Case Management
                                                  (pgvector)  & Admin Web Dashboard
```

- **Gateway & Backend**: Go (Chi HTTP Router, `pgxpool`, `go-redis`, Telegram Bot API)
- **AI Core & RAG**: Python 3.12 (OpenAI SDK, Pydantic Settings, `psycopg2`, `pgvector`)
- **Database**: PostgreSQL 16 + `pgvector`
- **Queue & Multi-turn Cache**: Redis 7 (Pub/Sub & Chat History List with 24h TTL)

---

## 🚀 Quick Start (Production Setup with Docker)

### 1. Prerequisites
- [Docker](https://www.docker.com/) & Docker Compose installed.

### 2. Environment Configuration
Copy `.env.example` to `.env` and configure your API Keys:
```bash
cp .env.example .env
```
Example `.env`:
```env
DATABASE_URL=postgres://user:password@localhost:5433/agenticsdk?sslmode=disable
REDIS_URL=redis://localhost:6380
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
LLM_API_KEY=sk-or-v1-your-openrouter-or-openai-key
LLM_BASE_URL=https://openrouter.ai/api/v1
LLM_MODEL=google/gemini-2.5-flash:free
OPENAI_EMBEDDING_API_KEY=sk-or-v1-your-key
OPENAI_EMBEDDING_BASE_URL=https://openrouter.ai/api/v1
```

### 3. Run the Entire System with Docker Compose
```bash
docker-compose up --build -d
```

### 4. Seed Database & Embed SOP Documents
```bash
# Seed initial tables
docker exec -i agentic_postgres psql -U user -d agenticsdk < db/migrations/001_create_users.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < db/migrations/002_create_docs.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < db/migrations/003_create_orders.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < db/migrations/004_create_cases.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < scripts/seed_users.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < scripts/seed_docs.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < scripts/seed_orders.sql

# Run SOP Vector Embedder
docker exec -it agentic_python_agent python -m app.rag.loader
```

---

## 💻 Web Dashboard & Testing

- **CS & Monitoring Dashboard**: Open [http://localhost:8080/dashboard](http://localhost:8080/dashboard) in your browser.
- **Run Integration Test Suite**:
  ```bash
  python scripts/test_e2e.py
  ```
