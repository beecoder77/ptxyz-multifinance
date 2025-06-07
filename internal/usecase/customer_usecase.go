package usecase

import (
	"errors"
	"sync"
	"time"
	"xyz-multifinance/internal/domain"
)

type customerUseCase struct {
	customerRepo domain.CustomerRepository
	mutex        sync.Mutex // For handling concurrent credit limit updates
}

// NewCustomerUseCase creates a new instance of CustomerUseCase
func NewCustomerUseCase(customerRepo domain.CustomerRepository) domain.CustomerUseCase {
	return &customerUseCase{
		customerRepo: customerRepo,
	}
}

// Register implements CustomerUseCase.Register
func (uc *customerUseCase) Register(customer *domain.Customer) error {
	// Check if customer with same NIK already exists
	existing, err := uc.customerRepo.GetByNIK(customer.NIK)
	if err != nil {
		if err.Error() != "not found" {
			return err
		}
	}

	if existing != nil {
		return errors.New("customer with this NIK already exists")
	}

	// Set timestamps
	now := time.Now()
	customer.CreatedAt = now
	customer.UpdatedAt = now

	// Create customer
	return uc.customerRepo.Create(customer)
}

// GetProfile implements CustomerUseCase.GetProfile
func (uc *customerUseCase) GetProfile(id uint) (*domain.Customer, error) {
	return uc.customerRepo.GetByID(id)
}

// UpdateProfile implements CustomerUseCase.UpdateProfile
func (uc *customerUseCase) UpdateProfile(customer *domain.Customer) error {
	existing, err := uc.customerRepo.GetByID(customer.ID)
	if err != nil {
		return err
	}

	// Update only allowed fields
	existing.FullName = customer.FullName
	existing.LegalName = customer.LegalName
	existing.Salary = customer.Salary
	existing.UpdatedAt = time.Now()

	return uc.customerRepo.Update(existing)
}

// GetCreditLimits implements CustomerUseCase.GetCreditLimits
func (uc *customerUseCase) GetCreditLimits(customerID uint) ([]domain.CreditLimit, error) {
	return uc.customerRepo.GetCreditLimits(customerID)
}

// CheckCreditLimit implements CustomerUseCase.CheckCreditLimit
func (uc *customerUseCase) CheckCreditLimit(customerID uint, amount float64, tenor int) (bool, error) {
	limits, err := uc.customerRepo.GetCreditLimits(customerID)
	if err != nil {
		return false, err
	}

	for _, limit := range limits {
		if limit.Tenor == tenor {
			return limit.GetAvailableLimit() >= amount, nil
		}
	}

	return false, errors.New("no credit limit found for the specified tenor")
}

// UpdateCreditLimitUsage implements CustomerUseCase.UpdateCreditLimitUsage
func (uc *customerUseCase) UpdateCreditLimitUsage(customerID uint, amount float64, tenor int) error {
	// Use mutex to prevent race conditions when updating credit limit
	uc.mutex.Lock()
	defer uc.mutex.Unlock()

	limits, err := uc.customerRepo.GetCreditLimits(customerID)
	if err != nil {
		return err
	}

	for _, limit := range limits {
		if limit.Tenor == tenor {
			if limit.GetAvailableLimit() < amount {
				return errors.New("insufficient credit limit")
			}

			limit.UsedAmount += amount
			limit.UpdatedAt = time.Now()
			return uc.customerRepo.UpdateCreditLimit(&limit)
		}
	}

	return errors.New("no credit limit found for the specified tenor")
}
