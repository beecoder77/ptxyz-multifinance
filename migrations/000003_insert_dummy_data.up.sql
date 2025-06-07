-- Insert dummy customers
INSERT INTO customers (
    nik, full_name, legal_name, place_of_birth, date_of_birth, 
    salary, ktp_photo, selfie_photo, created_at, updated_at
) VALUES 
    ('1234567890123456', 'John Doe', 'John Doe', 'Jakarta', '1990-01-01', 
    5000000, 'https://example.com/ktp1.jpg', 'https://example.com/selfie1.jpg', 
    NOW(), NOW()),
    ('2345678901234567', 'Jane Smith', 'Jane Smith', 'Surabaya', '1992-05-15', 
    7500000, 'https://example.com/ktp2.jpg', 'https://example.com/selfie2.jpg', 
    NOW(), NOW()),
    ('3456789012345678', 'Bob Wilson', 'Bob Wilson', 'Bandung', '1988-12-25', 
    10000000, 'https://example.com/ktp3.jpg', 'https://example.com/selfie3.jpg', 
    NOW(), NOW());

-- Insert credit limits for customers
INSERT INTO credit_limits (
    customer_id, tenor, amount, used_amount, created_at, updated_at
) VALUES 
    (1, 2, 10000000, 0, NOW(), NOW()),
    (1, 4, 20000000, 0, NOW(), NOW()),
    (2, 2, 15000000, 0, NOW(), NOW()),
    (2, 4, 30000000, 0, NOW(), NOW()),
    (3, 2, 20000000, 0, NOW(), NOW()),
    (3, 4, 40000000, 0, NOW(), NOW());

-- Insert dummy transactions
INSERT INTO transactions (
    customer_id, contract_number, source, asset_name, otr_amount,
    admin_fee, installment_amount, interest_amount, tenor,
    status, created_at, updated_at
) VALUES 
    (1, 'XYZ-1-20240308', 'e-commerce', 'Smartphone XYZ', 5000000,
    100000, 2583334, 500000, 2, 'approved', NOW(), NOW()),
    (2, 'XYZ-2-20240308', 'website', 'Laptop ABC', 12000000,
    150000, 6600000, 1200000, 2, 'pending', NOW(), NOW()),
    (3, 'XYZ-3-20240308', 'dealer', 'Camera DEF', 8000000,
    120000, 4400000, 800000, 2, 'approved', NOW(), NOW());

-- Insert installments for approved transactions
INSERT INTO installments (
    transaction_id, due_date, amount, status, paid_at, created_at, updated_at
) VALUES 
-- For Transaction 1 (2 months)
    (1, NOW() + INTERVAL '1 month', 2583334, 'paid', NOW(), NOW(), NOW()),
    (1, NOW() + INTERVAL '2 month', 2583334, 'unpaid', NULL, NOW(), NOW()),
-- For Transaction 3 (2 months)
    (3, NOW() + INTERVAL '1 month', 4400000, 'paid', NOW(), NOW(), NOW()),
    (3, NOW() + INTERVAL '2 month', 4400000, 'unpaid', NULL, NOW(), NOW()); 