-- Delete all data in reverse order to avoid foreign key constraints
DELETE FROM installments;
DELETE FROM transactions;
DELETE FROM credit_limits;
DELETE FROM customers; 