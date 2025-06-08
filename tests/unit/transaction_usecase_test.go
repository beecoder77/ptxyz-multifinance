package tests

import (
	"testing"
	"xyz-multifinance/internal/domain"
	"xyz-multifinance/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(tx *domain.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByID(id uint) (*domain.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByContractNumber(contractNumber string) (*domain.Transaction, error) {
	args := m.Called(contractNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(tx *domain.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockTransactionRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTransactionRepository) List(customerID uint, offset, limit int) ([]domain.Transaction, error) {
	args := m.Called(customerID, offset, limit)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetInstallments(transactionID uint) ([]domain.Installment, error) {
	args := m.Called(transactionID)
	return args.Get(0).([]domain.Installment), args.Error(1)
}

func (m *MockTransactionRepository) UpdateInstallment(installment *domain.Installment) error {
	args := m.Called(installment)
	return args.Error(0)
}

func TestTransactionUseCase_Create(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	mockCustomerUseCase := new(MockCustomerUseCase)

	useCase := usecase.NewTransactionUseCase(mockRepo, mockCustomerUseCase, nil)

	t.Run("Success", func(t *testing.T) {
		tx := &domain.Transaction{
			CustomerID:        1,
			Source:            domain.SourceECommerce,
			AssetName:         "Laptop",
			OTRAmount:         10000000,
			AdminFee:          100000,
			InstallmentAmount: 916667,
			InterestAmount:    1000000,
			Tenor:             12,
		}

		// Mock credit limit check
		totalAmount := tx.OTRAmount + tx.AdminFee
		mockCustomerUseCase.On("CheckCreditLimit", tx.CustomerID, totalAmount, tx.Tenor).Return(true, nil)

		// Mock transaction creation
		mockRepo.On("Create", mock.MatchedBy(func(t *domain.Transaction) bool {
			return t.CustomerID == tx.CustomerID &&
				t.Source == tx.Source &&
				t.AssetName == tx.AssetName &&
				t.OTRAmount == tx.OTRAmount &&
				t.AdminFee == tx.AdminFee &&
				t.InstallmentAmount == tx.InstallmentAmount &&
				t.InterestAmount == tx.InterestAmount &&
				t.Tenor == tx.Tenor
		})).Return(nil)

		// Mock credit limit update
		mockCustomerUseCase.On("UpdateCreditLimitUsage", tx.CustomerID, totalAmount, tx.Tenor).Return(nil)

		err := useCase.Create(tx)

		assert.NoError(t, err)
		assert.NotEmpty(t, tx.ContractNumber)
		mockRepo.AssertExpectations(t)
		mockCustomerUseCase.AssertExpectations(t)
	})
}
