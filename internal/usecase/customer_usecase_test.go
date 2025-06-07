package usecase

import (
	"errors"
	"testing"
	"time"
	"xyz-multifinance/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) Create(customer *domain.Customer) error {
	args := m.Called(customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) GetByID(id uint) (*domain.Customer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByNIK(nik string) (*domain.Customer, error) {
	args := m.Called(nik)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Update(customer *domain.Customer) error {
	args := m.Called(customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCustomerRepository) List(offset, limit int) ([]domain.Customer, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]domain.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetCreditLimits(customerID uint) ([]domain.CreditLimit, error) {
	args := m.Called(customerID)
	return args.Get(0).([]domain.CreditLimit), args.Error(1)
}

func (m *MockCustomerRepository) UpdateCreditLimit(limit *domain.CreditLimit) error {
	args := m.Called(limit)
	return args.Error(0)
}

func TestCustomerUseCase_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockCustomerRepository)
		useCase := NewCustomerUseCase(mockRepo)

		customer := &domain.Customer{
			NIK:          "1234567890123456",
			FullName:     "John Doe",
			LegalName:    "John Doe",
			PlaceOfBirth: "Jakarta",
			DateOfBirth:  time.Now().AddDate(-30, 0, 0),
			Salary:       5000000,
			KTPPhoto:     "ktp.jpg",
			SelfiePhoto:  "selfie.jpg",
		}

		mockRepo.On("GetByNIK", customer.NIK).Return(nil, errors.New("not found"))
		mockRepo.On("Create", mock.AnythingOfType("*domain.Customer")).Return(nil)

		err := useCase.Register(customer)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NIK Already Exists", func(t *testing.T) {
		mockRepo := new(MockCustomerRepository)
		useCase := NewCustomerUseCase(mockRepo)

		customer := &domain.Customer{
			NIK:          "1234567890123456",
			FullName:     "John Doe",
			LegalName:    "John Doe",
			PlaceOfBirth: "Jakarta",
			DateOfBirth:  time.Now().AddDate(-30, 0, 0),
			Salary:       5000000,
			KTPPhoto:     "ktp.jpg",
			SelfiePhoto:  "selfie.jpg",
		}

		existingCustomer := &domain.Customer{
			ID:           1,
			NIK:          "1234567890123456",
			FullName:     "Existing User",
			LegalName:    "Existing User",
			PlaceOfBirth: "Jakarta",
			DateOfBirth:  time.Now().AddDate(-25, 0, 0),
		}

		// Setup mock to return existing customer
		mockRepo.On("GetByNIK", customer.NIK).Return(existingCustomer, nil).Once()

		err := useCase.Register(customer)

		assert.Error(t, err)
		assert.Equal(t, "customer with this NIK already exists", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerUseCase_UpdateCreditLimitUsage(t *testing.T) {
	mockRepo := new(MockCustomerRepository)
	useCase := NewCustomerUseCase(mockRepo)

	t.Run("Success", func(t *testing.T) {
		customerID := uint(1)
		amount := float64(1000000)
		tenor := 12

		limits := []domain.CreditLimit{
			{
				CustomerID: customerID,
				Tenor:      tenor,
				Amount:     2000000,
				UsedAmount: 500000,
			},
		}

		mockRepo.On("GetCreditLimits", customerID).Return(limits, nil)
		mockRepo.On("UpdateCreditLimit", mock.AnythingOfType("*domain.CreditLimit")).Return(nil)

		err := useCase.UpdateCreditLimitUsage(customerID, amount, tenor)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Insufficient Credit Limit", func(t *testing.T) {
		customerID := uint(1)
		amount := float64(2000000)
		tenor := 12

		limits := []domain.CreditLimit{
			{
				CustomerID: customerID,
				Tenor:      tenor,
				Amount:     2000000,
				UsedAmount: 500000,
			},
		}

		mockRepo.On("GetCreditLimits", customerID).Return(limits, nil)

		err := useCase.UpdateCreditLimitUsage(customerID, amount, tenor)

		assert.Error(t, err)
		assert.Equal(t, "insufficient credit limit", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerUseCase_CheckCreditLimit(t *testing.T) {
	mockRepo := new(MockCustomerRepository)
	useCase := NewCustomerUseCase(mockRepo)

	t.Run("Has Sufficient Limit", func(t *testing.T) {
		customerID := uint(1)
		amount := float64(1000000)
		tenor := 12

		limits := []domain.CreditLimit{
			{
				CustomerID: customerID,
				Tenor:      tenor,
				Amount:     2000000,
				UsedAmount: 500000,
			},
		}

		mockRepo.On("GetCreditLimits", customerID).Return(limits, nil)

		hasLimit, err := useCase.CheckCreditLimit(customerID, amount, tenor)

		assert.NoError(t, err)
		assert.True(t, hasLimit)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Insufficient Limit", func(t *testing.T) {
		customerID := uint(1)
		amount := float64(2000000)
		tenor := 12

		limits := []domain.CreditLimit{
			{
				CustomerID: customerID,
				Tenor:      tenor,
				Amount:     2000000,
				UsedAmount: 500000,
			},
		}

		mockRepo.On("GetCreditLimits", customerID).Return(limits, nil)

		hasLimit, err := useCase.CheckCreditLimit(customerID, amount, tenor)

		assert.NoError(t, err)
		assert.False(t, hasLimit)
		mockRepo.AssertExpectations(t)
	})
}
