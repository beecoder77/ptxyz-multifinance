package repository

import (
	"errors"
	"xyz-multifinance/internal/domain"

	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new instance of TransactionRepository
func NewTransactionRepository(db *gorm.DB) domain.TransactionRepository {
	return &transactionRepository{
		db: db,
	}
}

// Create implements TransactionRepository.Create
func (r *transactionRepository) Create(tx *domain.Transaction) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create transaction
		if err := tx.Create(&tx).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetByID implements TransactionRepository.GetByID
func (r *transactionRepository) GetByID(id uint) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.db.Preload("Customer").Preload("Installments").First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// GetByContractNumber implements TransactionRepository.GetByContractNumber
func (r *transactionRepository) GetByContractNumber(contractNumber string) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.db.Preload("Customer").Preload("Installments").
		Where("contract_number = ?", contractNumber).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// Update implements TransactionRepository.Update
func (r *transactionRepository) Update(tx *domain.Transaction) error {
	return r.db.Transaction(func(db *gorm.DB) error {
		// Get current version
		var current domain.Transaction
		if err := db.Select("version").First(&current, tx.ID).Error; err != nil {
			return err
		}

		// Check version
		if current.Version != tx.Version {
			return errors.New("concurrent modification detected")
		}

		// Increment version
		tx.Version++

		// Update transaction
		if err := db.Save(tx).Error; err != nil {
			return err
		}

		return nil
	})
}

// Delete implements TransactionRepository.Delete
func (r *transactionRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Transaction{}, id).Error
}

// List implements TransactionRepository.List
func (r *transactionRepository) List(customerID uint, offset, limit int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.db.Preload("Installments").
		Where("customer_id = ?", customerID).
		Offset(offset).Limit(limit).
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetInstallments implements TransactionRepository.GetInstallments
func (r *transactionRepository) GetInstallments(transactionID uint) ([]domain.Installment, error) {
	var installments []domain.Installment
	err := r.db.Where("transaction_id = ?", transactionID).
		Order("due_date asc").
		Find(&installments).Error
	if err != nil {
		return nil, err
	}
	return installments, nil
}

// UpdateInstallment implements TransactionRepository.UpdateInstallment
func (r *transactionRepository) UpdateInstallment(installment *domain.Installment) error {
	return r.db.Transaction(func(db *gorm.DB) error {
		// Get current version
		var current domain.Installment
		if err := db.Select("version").First(&current, installment.ID).Error; err != nil {
			return err
		}

		// Check version
		if current.Version != installment.Version {
			return errors.New("concurrent modification detected")
		}

		// Increment version
		installment.Version++

		// Update installment
		if err := db.Save(installment).Error; err != nil {
			return err
		}

		return nil
	})
}
