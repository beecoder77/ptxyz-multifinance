package domain

import (
	"time"
)

// Customer represents the customer entity
type Customer struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	NIK          string     `json:"nik" gorm:"unique;not null"`
	FullName     string     `json:"full_name" gorm:"not null"`
	LegalName    string     `json:"legal_name" gorm:"not null"`
	PlaceOfBirth string     `json:"place_of_birth" gorm:"not null"`
	DateOfBirth  time.Time  `json:"date_of_birth" gorm:"not null"`
	Salary       float64    `json:"salary" gorm:"not null"`
	KTPPhoto     string     `json:"ktp_photo" gorm:"not null"`
	SelfiePhoto  string     `json:"selfie_photo" gorm:"not null"`
	Version      int        `json:"version" gorm:"not null;default:1"` // For optimistic locking
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	CreditLimits []CreditLimit `json:"credit_limits,omitempty" gorm:"foreignKey:CustomerID"`
	Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:CustomerID"`
}

// CreditLimit represents the credit limit for different tenors
type CreditLimit struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CustomerID uint      `json:"customer_id" gorm:"not null"`
	Tenor      int       `json:"tenor" gorm:"not null"` // in months
	Amount     float64   `json:"amount" gorm:"not null"`
	UsedAmount float64   `json:"used_amount" gorm:"not null;default:0"`
	Version    int       `json:"version" gorm:"not null;default:1"` // For optimistic locking
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GetAvailableLimit calculates remaining credit limit
func (cl *CreditLimit) GetAvailableLimit() float64 {
	return cl.Amount - cl.UsedAmount
}

// CustomerRepository represents the customer repository contract
type CustomerRepository interface {
	Create(customer *Customer) error
	GetByID(id uint) (*Customer, error)
	GetByNIK(nik string) (*Customer, error)
	Update(customer *Customer) error
	Delete(id uint) error
	List(offset, limit int) ([]Customer, error)
	GetCreditLimits(customerID uint) ([]CreditLimit, error)
	UpdateCreditLimit(limit *CreditLimit) error
}

// CustomerUseCase represents the customer use case contract
type CustomerUseCase interface {
	Register(customer *Customer) error
	GetProfile(id uint) (*Customer, error)
	UpdateProfile(customer *Customer) error
	GetCreditLimits(customerID uint) ([]CreditLimit, error)
	CheckCreditLimit(customerID uint, amount float64, tenor int) (bool, error)
	UpdateCreditLimitUsage(customerID uint, amount float64, tenor int) error
}
