package repository

import (
	"xyz-multifinance/internal/domain"

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
	return r.db.Save(customer).Error
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
	return r.db.Save(limit).Error
}
