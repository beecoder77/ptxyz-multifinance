package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"xyz-multifinance/internal/domain"
	"xyz-multifinance/internal/pkg/redis"
)

type transactionUseCase struct {
	transactionRepo domain.TransactionRepository
	customerUseCase domain.CustomerUseCase
	redisClient     redis.RedisClient
}

// NewTransactionUseCase creates a new instance of TransactionUseCase
func NewTransactionUseCase(
	transactionRepo domain.TransactionRepository,
	customerUseCase domain.CustomerUseCase,
	redisClient redis.RedisClient,
) domain.TransactionUseCase {
	return &transactionUseCase{
		transactionRepo: transactionRepo,
		customerUseCase: customerUseCase,
		redisClient:     redisClient,
	}
}

// Create implements TransactionUseCase.Create
func (uc *transactionUseCase) Create(tx *domain.Transaction) error {
	// Check credit limit
	totalAmount := tx.OTRAmount + tx.AdminFee
	hasLimit, err := uc.customerUseCase.CheckCreditLimit(tx.CustomerID, totalAmount, tx.Tenor)
	if err != nil {
		return err
	}
	if !hasLimit {
		return errors.New("insufficient credit limit")
	}

	// Generate contract number
	tx.ContractNumber = fmt.Sprintf("XYZ-%d-%d", tx.CustomerID, time.Now().Unix())
	tx.Status = domain.StatusPending
	tx.Version = 1 // Initialize version for optimistic locking

	// Set timestamps
	now := time.Now()
	tx.CreatedAt = now
	tx.UpdatedAt = now

	// Create installments
	monthlyInstallment := tx.InstallmentAmount
	for i := 0; i < tx.Tenor; i++ {
		dueDate := now.AddDate(0, i+1, 0) // Due date is next month from creation
		installment := domain.Installment{
			TransactionID: tx.ID,
			DueDate:       dueDate,
			Amount:        monthlyInstallment,
			Status:        "unpaid",
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		tx.Installments = append(tx.Installments, installment)
	}

	// Create transaction
	if err := uc.transactionRepo.Create(tx); err != nil {
		return err
	}

	// Update credit limit usage
	return uc.customerUseCase.UpdateCreditLimitUsage(tx.CustomerID, totalAmount, tx.Tenor)
}

// GetByID implements TransactionUseCase.GetByID
func (uc *transactionUseCase) GetByID(id uint) (*domain.Transaction, error) {
	return uc.transactionRepo.GetByID(id)
}

// GetByContractNumber implements TransactionUseCase.GetByContractNumber
func (uc *transactionUseCase) GetByContractNumber(contractNumber string) (*domain.Transaction, error) {
	return uc.transactionRepo.GetByContractNumber(contractNumber)
}

// UpdateStatus implements TransactionUseCase.UpdateStatus
func (uc *transactionUseCase) UpdateStatus(id uint, status domain.TransactionStatus) error {
	tx, err := uc.transactionRepo.GetByID(id)
	if err != nil {
		return err
	}

	tx.Status = status
	tx.UpdatedAt = time.Now()

	return uc.transactionRepo.Update(tx)
}

// GetCustomerTransactions implements TransactionUseCase.GetCustomerTransactions
func (uc *transactionUseCase) GetCustomerTransactions(customerID uint, offset, limit int) ([]domain.Transaction, error) {
	return uc.transactionRepo.List(customerID, offset, limit)
}

// GetInstallments implements TransactionUseCase.GetInstallments
func (uc *transactionUseCase) GetInstallments(transactionID uint) ([]domain.Installment, error) {
	return uc.transactionRepo.GetInstallments(transactionID)
}

// PayInstallment implements TransactionUseCase.PayInstallment
func (uc *transactionUseCase) PayInstallment(installmentID uint) error {
	// Create distributed lock
	lock := redis.NewDistributedLock(uc.redisClient, fmt.Sprintf("installment:%d", installmentID), 30*time.Second)

	// Try to acquire lock with timeout
	ctx := context.Background()
	if err := lock.TryLock(ctx, 5*time.Second); err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer lock.Unlock(ctx)

	// Get installment with retries for optimistic locking
	var targetInstallment *domain.Installment
	maxRetries := 3
	var lastError error

	for i := 0; i < maxRetries; i++ {
		installments, err := uc.transactionRepo.GetInstallments(installmentID)
		if err != nil {
			return err
		}

		for _, inst := range installments {
			if inst.ID == installmentID {
				targetInstallment = &inst
				break
			}
		}

		if targetInstallment == nil {
			return errors.New("installment not found")
		}

		if targetInstallment.Status == "paid" {
			return errors.New("installment already paid")
		}

		now := time.Now()
		targetInstallment.Status = "paid"
		targetInstallment.PaidAt = &now
		targetInstallment.UpdatedAt = now

		err = uc.transactionRepo.UpdateInstallment(targetInstallment)
		if err != nil {
			if err.Error() == "concurrent modification detected" {
				lastError = err
				time.Sleep(100 * time.Millisecond) // Wait before retry
				continue
			}
			return err
		}
		return nil
	}

	return fmt.Errorf("failed to update installment after %d retries: %v", maxRetries, lastError)
}
