INSERT INTO users (uuid, name, division) VALUES
    (gen_random_uuid(), 'Andi Pratama', 'payment'),
    (gen_random_uuid(), 'Siti Rahayu', 'shipping'),
    (gen_random_uuid(), 'Budi Santoso', 'IT')
ON CONFLICT DO NOTHING;
