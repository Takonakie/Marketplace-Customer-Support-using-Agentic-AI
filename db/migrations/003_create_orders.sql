CREATE TABLE IF NOT EXISTS orders (
    uuid         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id  VARCHAR(255) NOT NULL,    -- telegram user ID or internal customer ID
    product_name VARCHAR(255) NOT NULL,
    amount       BIGINT NOT NULL,          -- price in smallest currency unit (e.g. Rupiah)
    status       VARCHAR(50) NOT NULL,     -- "pending", "shipped", "delivered", "cancelled", "refunded"
    order_date   TIMESTAMP DEFAULT NOW(),
    tracking_id  VARCHAR(255),
    created_at   TIMESTAMP DEFAULT NOW(),
    updated_at   TIMESTAMP DEFAULT NOW()
);
