CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS docs (
    uuid       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,       -- e.g. "SOP Refund"
    doc        TEXT NOT NULL,               -- full document content
    embedding  VECTOR(1536),               -- vector embedding for RAG search
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
