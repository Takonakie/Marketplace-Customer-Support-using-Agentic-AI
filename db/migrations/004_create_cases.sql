CREATE TABLE IF NOT EXISTS cases (
    uuid        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,      -- short title of the case
    datetime    TIMESTAMP DEFAULT NOW(),    -- when the case was created
    criticality VARCHAR(50) NOT NULL,       -- "low", "medium", "high", "critical"
    description TEXT NOT NULL,              -- detailed description of the issue
    assign_to   UUID REFERENCES users(uuid),-- assigned staff member
    status      VARCHAR(50) DEFAULT 'open', -- "open", "in_progress", "resolved", "closed"
    resolution  TEXT,                       -- how the case was resolved (filled later)
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);
