package http

import (
	"net/http"
	"strconv"
	"xyz-multifinance/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TransactionHandler struct {
	transactionUseCase domain.TransactionUseCase
	validate           *validator.Validate
}

func NewTransactionHandler(router *gin.Engine, transactionUseCase domain.TransactionUseCase) {
	handler := &TransactionHandler{
		transactionUseCase: transactionUseCase,
		validate:           validator.New(),
	}

	transactionRoutes := router.Group("/api/v1/transactions")
	{
		transactionRoutes.POST("", handler.Create)
		transactionRoutes.GET("/:id", handler.GetByID)
		transactionRoutes.GET("/contract/:number", handler.GetByContractNumber)
		transactionRoutes.PUT("/:id/status", handler.UpdateStatus)
		transactionRoutes.GET("/customer/:customer_id", handler.GetCustomerTransactions)
		transactionRoutes.GET("/:id/installments", handler.GetInstallments)
		transactionRoutes.POST("/installments/:id/pay", handler.PayInstallment)
	}
}

type CreateTransactionRequest struct {
	CustomerID        uint                     `json:"customer_id" validate:"required"`
	Source            domain.TransactionSource `json:"source" validate:"required,oneof=e-commerce website dealer"`
	AssetName         string                   `json:"asset_name" validate:"required"`
	OTRAmount         float64                  `json:"otr_amount" validate:"required,gt=0"`
	AdminFee          float64                  `json:"admin_fee" validate:"required,gte=0"`
	InstallmentAmount float64                  `json:"installment_amount" validate:"required,gt=0"`
	InterestAmount    float64                  `json:"interest_amount" validate:"required,gte=0"`
	Tenor             int                      `json:"tenor" validate:"required,oneof=1 2 3 4"`
}

func (h *TransactionHandler) Create(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := &domain.Transaction{
		CustomerID:        req.CustomerID,
		Source:            req.Source,
		AssetName:         req.AssetName,
		OTRAmount:         req.OTRAmount,
		AdminFee:          req.AdminFee,
		InstallmentAmount: req.InstallmentAmount,
		InterestAmount:    req.InterestAmount,
		Tenor:             req.Tenor,
	}

	if err := h.transactionUseCase.Create(tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tx)
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	tx, err := h.transactionUseCase.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *TransactionHandler) GetByContractNumber(c *gin.Context) {
	number := c.Param("number")
	tx, err := h.transactionUseCase.GetByContractNumber(number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

type UpdateStatusRequest struct {
	Status domain.TransactionStatus `json:"status" validate:"required,oneof=pending approved rejected cancelled"`
}

func (h *TransactionHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.transactionUseCase.UpdateStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

func (h *TransactionHandler) GetCustomerTransactions(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("customer_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	transactions, err := h.transactionUseCase.GetCustomerTransactions(uint(customerID), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func (h *TransactionHandler) GetInstallments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	installments, err := h.transactionUseCase.GetInstallments(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, installments)
}

func (h *TransactionHandler) PayInstallment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid installment ID"})
		return
	}

	if err := h.transactionUseCase.PayInstallment(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "installment paid successfully"})
}
