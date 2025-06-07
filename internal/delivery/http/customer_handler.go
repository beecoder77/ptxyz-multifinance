package http

import (
	"net/http"
	"strconv"
	"time"
	"xyz-multifinance/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CustomerHandler struct {
	customerUseCase domain.CustomerUseCase
	validate        *validator.Validate
}

func NewCustomerHandler(router *gin.Engine, customerUseCase domain.CustomerUseCase) {
	handler := &CustomerHandler{
		customerUseCase: customerUseCase,
		validate:        validator.New(),
	}

	customerRoutes := router.Group("/api/v1/customers")
	{
		customerRoutes.POST("", handler.Register)
		customerRoutes.GET("/:id", handler.GetProfile)
		customerRoutes.PUT("/:id", handler.UpdateProfile)
		customerRoutes.GET("/:id/credit-limits", handler.GetCreditLimits)
	}
}

type RegisterRequest struct {
	NIK          string  `json:"nik" validate:"required,len=16"`
	FullName     string  `json:"full_name" validate:"required"`
	LegalName    string  `json:"legal_name" validate:"required"`
	PlaceOfBirth string  `json:"place_of_birth" validate:"required"`
	DateOfBirth  string  `json:"date_of_birth" validate:"required,datetime=2006-01-02"`
	Salary       float64 `json:"salary" validate:"required,gt=0"`
	KTPPhoto     string  `json:"ktp_photo" validate:"required,url"`
	SelfiePhoto  string  `json:"selfie_photo" validate:"required,url"`
}

func (h *CustomerHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse date of birth
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	customer := &domain.Customer{
		NIK:          req.NIK,
		FullName:     req.FullName,
		LegalName:    req.LegalName,
		PlaceOfBirth: req.PlaceOfBirth,
		DateOfBirth:  dob,
		Salary:       req.Salary,
		KTPPhoto:     req.KTPPhoto,
		SelfiePhoto:  req.SelfiePhoto,
	}

	if err := h.customerUseCase.Register(customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

func (h *CustomerHandler) GetProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	customer, err := h.customerUseCase.GetProfile(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

type UpdateProfileRequest struct {
	FullName  string  `json:"full_name" validate:"required"`
	LegalName string  `json:"legal_name" validate:"required"`
	Salary    float64 `json:"salary" validate:"required,gt=0"`
}

func (h *CustomerHandler) UpdateProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer := &domain.Customer{
		ID:        uint(id),
		FullName:  req.FullName,
		LegalName: req.LegalName,
		Salary:    req.Salary,
	}

	if err := h.customerUseCase.UpdateProfile(customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *CustomerHandler) GetCreditLimits(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	limits, err := h.customerUseCase.GetCreditLimits(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, limits)
}
