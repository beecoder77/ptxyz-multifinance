package repository

import (
	"errors"
	"fmt"
	"time"
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
func (r *transactionRepository) Create(transaction *domain.Transaction) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Set initial version
		transaction.Version = 1

		// Create transaction with specific column order using raw SQL
		result := tx.Raw(`INSERT INTO "transactions" ("customer_id","contract_number","source","status","asset_name","otr_amount","admin_fee","installment_amount","interest_amount","tenor","version","created_at","updated_at","deleted_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING "id"`,
			transaction.CustomerID, transaction.ContractNumber,
			transaction.Source, transaction.Status, transaction.AssetName,
			transaction.OTRAmount, transaction.AdminFee,
			transaction.InstallmentAmount, transaction.InterestAmount,
			transaction.Tenor, transaction.Version,
			time.Now(), time.Now(), nil,
		).Scan(&transaction.ID)

		if result.Error != nil {
			return result.Error
		}

		// Create installments with specific column order using raw SQL
		for i := 1; i <= transaction.Tenor; i++ {
			dueDate := time.Now().AddDate(0, i, 0)
			result := tx.Raw(`INSERT INTO "installments" ("transaction_id","installment_number","amount","status","due_date","version","created_at","updated_at","deleted_at") VALUES (?,?,?,?,?,?,?,?,?) RETURNING "id"`,
				transaction.ID, i, transaction.InstallmentAmount, "unpaid",
				dueDate, 1, time.Now(), time.Now(), nil,
			).Scan(new(uint))

			if result.Error != nil {
				return result.Error
			}
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

	// Get transaction
	query := `SELECT * FROM "transactions" WHERE "contract_number" = ? AND "deleted_at" IS NULL`
	if err := r.db.Raw(query, contractNumber).Scan(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Get customer
	var customer domain.Customer
	customerQuery := `SELECT * FROM "customers" WHERE "id" = ? AND "deleted_at" IS NULL`
	if err := r.db.Raw(customerQuery, transaction.CustomerID).Scan(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	transaction.Customer = &customer

	// Get installments
	var installments []domain.Installment
	installmentQuery := `SELECT * FROM "installments" WHERE "transaction_id" = ? AND "deleted_at" IS NULL ORDER BY "installment_number" ASC`
	if err := r.db.Raw(installmentQuery, transaction.ID).Scan(&installments).Error; err != nil {
		return nil, fmt.Errorf("failed to get installments: %w", err)
	}
	transaction.Installments = installments

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
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Increment version
		installment.Version++

		// Update installment using raw SQL
		result := tx.Exec(`UPDATE "installments" SET "transaction_id"=?,"installment_number"=?,"amount"=?,"status"=?,"due_date"=?,"updated_at"=?,"version"=? WHERE "id"=? AND "version"=?`,
			installment.TransactionID,
			installment.InstallmentNumber,
			installment.Amount,
			installment.Status,
			installment.DueDate,
			time.Now(),
			installment.Version,
			installment.ID,
			installment.Version-1,
		)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("optimistic lock error")
		}

		return nil
	})
}
