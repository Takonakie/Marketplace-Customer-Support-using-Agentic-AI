#!/bin/bash
set -e

echo "=== 🚀 Starting Automated Setup & Seeding Script ==="

echo "1. Applying Database Migrations..."
docker exec -i agentic_postgres psql -U user -d agenticsdk < db/migrations/001_create_users.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < db/migrations/002_create_docs.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < db/migrations/003_create_orders.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < db/migrations/004_create_cases.sql

echo "2. Seeding Initial Data..."
docker exec -i agentic_postgres psql -U user -d agenticsdk < scripts/seed_users.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < scripts/seed_docs.sql
docker exec -i agentic_postgres psql -U user -d agenticsdk < scripts/seed_orders.sql

echo "3. Running SOP Vector Embedder..."
docker exec -i agentic_python_agent python -m app.rag.loader

echo "=== ✅ All Setup & Seeding Steps Completed Successfully! ==="
