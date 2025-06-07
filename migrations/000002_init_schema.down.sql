-- Drop triggers
DROP TRIGGER IF EXISTS update_installments_updated_at ON installments;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP TRIGGER IF EXISTS update_credit_limits_updated_at ON credit_limits;
DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_installments_due_date;
DROP INDEX IF EXISTS idx_installments_transaction_id;
DROP INDEX IF EXISTS idx_transactions_customer_id;
DROP INDEX IF EXISTS idx_transactions_contract_number;
DROP INDEX IF EXISTS idx_credit_limits_customer_id;
DROP INDEX IF EXISTS idx_customers_nik;

-- Drop tables
DROP TABLE IF EXISTS installments;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS credit_limits;
DROP TABLE IF EXISTS customers; 