package repository

import (
	"xyz-multifinance/internal/domain"

	"errors"
	"time"

	"gorm.io/gorm"
)

type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new instance of CustomerRepository
func NewCustomerRepository(db *gorm.DB) domain.CustomerRepository {
	return &customerRepository{
		db: db,
	}
}

// Create implements CustomerRepository.Create
func (r *customerRepository) Create(customer *domain.Customer) error {
	return r.db.Create(customer).Error
}

// GetByID implements CustomerRepository.GetByID
func (r *customerRepository) GetByID(id uint) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.Preload("CreditLimits").First(&customer, id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// GetByNIK implements CustomerRepository.GetByNIK
func (r *customerRepository) GetByNIK(nik string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.Preload("CreditLimits").Where("nik = ?", nik).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// Update implements CustomerRepository.Update
func (r *customerRepository) Update(customer *domain.Customer) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get current version using raw SQL
		var current struct {
			Version int
		}
		if err := tx.Raw(`SELECT version FROM "customers" WHERE "customers"."id" = ? ORDER BY "customers"."id" LIMIT ?`, customer.ID, 1).Scan(&current).Error; err != nil {
			return err
		}

		// Check version
		if current.Version != customer.Version {
			return errors.New("concurrent modification detected")
		}

		// Increment version
		customer.Version++

		// Update customer
		if err := tx.Save(customer).Error; err != nil {
			return err
		}

		return nil
	})
}

// Delete implements CustomerRepository.Delete
func (r *customerRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Customer{}, id).Error
}

// List implements CustomerRepository.List
func (r *customerRepository) List(offset, limit int) ([]domain.Customer, error) {
	var customers []domain.Customer
	err := r.db.Offset(offset).Limit(limit).Find(&customers).Error
	if err != nil {
		return nil, err
	}
	return customers, nil
}

// GetCreditLimits implements CustomerRepository.GetCreditLimits
func (r *customerRepository) GetCreditLimits(customerID uint) ([]domain.CreditLimit, error) {
	var limits []domain.CreditLimit
	err := r.db.Where("customer_id = ?", customerID).Find(&limits).Error
	if err != nil {
		return nil, err
	}
	return limits, nil
}

// UpdateCreditLimit implements CustomerRepository.UpdateCreditLimit
func (r *customerRepository) UpdateCreditLimit(limit *domain.CreditLimit) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get current version using raw SQL
		var current struct {
			Version int
		}
		if err := tx.Raw(`SELECT version FROM "credit_limits" WHERE "credit_limits"."id" = ? ORDER BY "credit_limits"."id" LIMIT ?`, limit.ID, 1).Scan(&current).Error; err != nil {
			return err
		}

		// Check version
		if current.Version != limit.Version {
			return errors.New("concurrent modification detected")
		}

		// Increment version
		limit.Version++

		// Update credit limit using raw SQL
		result := tx.Exec(`UPDATE "credit_limits" SET "customer_id"=?,"tenor"=?,"amount"=?,"used_amount"=?,"version"=?,"created_at"=?,"updated_at"=?,"deleted_at"=? WHERE "id" = ?`,
			limit.CustomerID, limit.Tenor, limit.Amount, limit.UsedAmount,
			limit.Version, time.Time{}, time.Now(), nil,
			limit.ID,
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
