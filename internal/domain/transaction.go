package domain

import (
	"time"
)

// TransactionSource represents the source of transaction
type TransactionSource string

const (
	SourceECommerce TransactionSource = "e-commerce"
	SourceWebsite   TransactionSource = "website"
	SourceDealer    TransactionSource = "dealer"
)

// TransactionStatus represents the status of transaction
type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusApproved  TransactionStatus = "approved"
	StatusRejected  TransactionStatus = "rejected"
	StatusCancelled TransactionStatus = "cancelled"
)

// Transaction represents the transaction entity
type Transaction struct {
	ID                uint              `json:"id" gorm:"primaryKey"`
	ContractNumber    string            `json:"contract_number" gorm:"unique;not null"`
	CustomerID        uint              `json:"customer_id" gorm:"not null"`
	Source            TransactionSource `json:"source" gorm:"not null"`
	Status            TransactionStatus `json:"status" gorm:"not null"`
	AssetName         string            `json:"asset_name" gorm:"not null"`
	OTRAmount         float64           `json:"otr_amount" gorm:"not null"` // On The Road price
	AdminFee          float64           `json:"admin_fee" gorm:"not null"`
	InstallmentAmount float64           `json:"installment_amount" gorm:"not null"`
	InterestAmount    float64           `json:"interest_amount" gorm:"not null"`
	Tenor             int               `json:"tenor" gorm:"not null"`             // in months
	Version           int               `json:"version" gorm:"not null;default:1"` // For optimistic locking
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DeletedAt         *time.Time        `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Customer     *Customer     `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Installments []Installment `json:"installments,omitempty" gorm:"foreignKey:TransactionID"`
}

// Installment represents the installment entity
type Installment struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	TransactionID uint       `json:"transaction_id" gorm:"not null"`
	DueDate       time.Time  `json:"due_date" gorm:"not null"`
	Amount        float64    `json:"amount" gorm:"not null"`
	Status        string     `json:"status" gorm:"not null;default:'unpaid'"` // paid, unpaid, overdue
	Version       int        `json:"version" gorm:"not null;default:1"`       // For optimistic locking
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TransactionRepository represents the transaction repository contract
type TransactionRepository interface {
	Create(tx *Transaction) error
	GetByID(id uint) (*Transaction, error)
	GetByContractNumber(contractNumber string) (*Transaction, error)
	Update(tx *Transaction) error
	Delete(id uint) error
	List(customerID uint, offset, limit int) ([]Transaction, error)
	GetInstallments(transactionID uint) ([]Installment, error)
	UpdateInstallment(installment *Installment) error
}

// TransactionUseCase represents the transaction use case contract
type TransactionUseCase interface {
	Create(tx *Transaction) error
	GetByID(id uint) (*Transaction, error)
	GetByContractNumber(contractNumber string) (*Transaction, error)
	UpdateStatus(id uint, status TransactionStatus) error
	GetCustomerTransactions(customerID uint, offset, limit int) ([]Transaction, error)
	GetInstallments(transactionID uint) ([]Installment, error)
	PayInstallment(installmentID uint) error
}
