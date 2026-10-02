INSERT INTO orders (uuid, customer_id, product_name, amount, status, tracking_id) VALUES
    (gen_random_uuid(), 'tg_12345', 'Raket Yonex Astrox 88D', 1299100, 'shipped', 'JNE-1828772'),
    (gen_random_uuid(), 'tg_12345', 'Sepatu Badminton Li-Ning', 899000, 'delivered', 'JNE-1828001'),
    (gen_random_uuid(), 'tg_67890', 'Shuttlecock Mavis 350', 150000, 'cancelled', NULL)
ON CONFLICT DO NOTHING;
