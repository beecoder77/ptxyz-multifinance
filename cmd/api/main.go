package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpHandler "xyz-multifinance/internal/delivery/http"
	"xyz-multifinance/internal/middleware"
	"xyz-multifinance/internal/pkg/crypto"
	"xyz-multifinance/internal/repository"
	"xyz-multifinance/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Load configuration
	loadConfig()

	// Initialize encryption
	if err := crypto.InitEncryption(viper.GetString("security.encryption_key")); err != nil {
		sugar.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Initialize database
	db, err := initDB()
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Redis
	redisClient := initRedis()

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Initialize use cases
	customerUseCase := usecase.NewCustomerUseCase(customerRepo)
	transactionUseCase := usecase.NewTransactionUseCase(transactionRepo, customerUseCase, redisClient)

	// Initialize Gin router
	router := gin.Default()

	// Initialize middlewares
	authConfig := middleware.AuthConfig{
		SecretKey: viper.GetString("jwt.secret"),
		Issuer:    viper.GetString("jwt.issuer"),
	}
	rateLimiterConfig := middleware.RateLimiterConfig{
		RedisClient: redisClient,
		MaxRequests: viper.GetInt("rate_limit.max_requests"),
		Window:      time.Duration(viper.GetInt("rate_limit.window")) * time.Second,
	}

	// Apply global middlewares
	router.Use(
		middleware.SecurityHeadersMiddleware(),
		middleware.NewSQLInjectionMiddleware(),
		middleware.NewRateLimiterMiddleware(rateLimiterConfig),
	)

	// Initialize HTTP handlers
	httpHandler.NewCustomerHandler(router, customerUseCase)
	httpHandler.NewTransactionHandler(router, transactionUseCase)

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.NewAuthMiddleware(authConfig))

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("server.port")),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		sugar.Fatalf("Server forced to shutdown: %v", err)
	}

	sugar.Info("Server exiting")
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}

func initDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.name"),
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})
}
