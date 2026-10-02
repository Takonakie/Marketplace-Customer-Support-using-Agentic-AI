CREATE TABLE IF NOT EXISTS users (
    uuid     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name     VARCHAR(255) NOT NULL,
    division VARCHAR(50) NOT NULL          -- "payment", "shipping", "IT"
);
