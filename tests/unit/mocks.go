package tests

import (
	"context"
	"time"
	"xyz-multifinance/internal/domain"
	"xyz-multifinance/internal/pkg/redis"

	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

// MockCustomerUseCase is a mock for CustomerUseCase interface
type MockCustomerUseCase struct {
	mock.Mock
}

func (m *MockCustomerUseCase) Register(customer *domain.Customer) error {
	args := m.Called(customer)
	return args.Error(0)
}

func (m *MockCustomerUseCase) GetProfile(id uint) (*domain.Customer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCustomerUseCase) UpdateProfile(customer *domain.Customer) error {
	args := m.Called(customer)
	return args.Error(0)
}

func (m *MockCustomerUseCase) GetCreditLimits(customerID uint) ([]domain.CreditLimit, error) {
	args := m.Called(customerID)
	return args.Get(0).([]domain.CreditLimit), args.Error(1)
}

func (m *MockCustomerUseCase) CheckCreditLimit(customerID uint, amount float64, tenor int) (bool, error) {
	args := m.Called(customerID, amount, tenor)
	return args.Bool(0), args.Error(1)
}

func (m *MockCustomerUseCase) UpdateCreditLimitUsage(customerID uint, amount float64, tenor int) error {
	args := m.Called(customerID, amount, tenor)
	return args.Error(0)
}

// Ensure MockRedisClient implements redis.RedisClient interface
var _ redis.RedisClient = (*MockRedisClient)(nil)

// MockRedisClient is a mock for Redis client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redisClient.BoolCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redisClient.BoolCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redisClient.IntCmd {
	args := m.Called(ctx, keys[0])
	return args.Get(0).(*redisClient.IntCmd)
}

func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redisClient.Cmd {
	mockArgs := m.Called(ctx, script, keys, args)
	if cmd, ok := mockArgs.Get(0).(*redisClient.IntCmd); ok {
		result := redisClient.NewCmd(ctx)
		result.SetVal(cmd.Val())
		return result
	}
	return mockArgs.Get(0).(*redisClient.Cmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redisClient.StringCmd {
	args := m.Called(ctx, key)
	if cmd, ok := args.Get(0).(*redisClient.StringCmd); ok {
		return cmd
	}
	return redisClient.NewStringResult(args.String(0), args.Error(1))
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redisClient.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return redisClient.NewStatusResult(args.String(0), args.Error(1))
}

// Implement other required methods for redis.Cmdable
func (m *MockRedisClient) Pipeline() redisClient.Pipeliner {
	args := m.Called()
	return args.Get(0).(redisClient.Pipeliner)
}

func (m *MockRedisClient) TxPipeline() redisClient.Pipeliner {
	args := m.Called()
	return args.Get(0).(redisClient.Pipeliner)
}

func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) *redisClient.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redisClient.PubSub)
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redisClient.IntCmd {
	args := m.Called(ctx, channel, message)
	return redisClient.NewIntResult(args.Get(0).(int64), args.Error(1))
}

func (m *MockRedisClient) Watch(ctx context.Context, fn func(*redisClient.Tx) error, keys ...string) error {
	args := m.Called(ctx, fn, keys)
	return args.Error(0)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Additional methods required by redis.Client
func (m *MockRedisClient) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockRedisClient) Options() *redisClient.Options {
	args := m.Called()
	return args.Get(0).(*redisClient.Options)
}

func (m *MockRedisClient) AddHook(hook redisClient.Hook) {
	m.Called(hook)
}

func (m *MockRedisClient) Conn() *redisClient.Conn {
	args := m.Called()
	return args.Get(0).(*redisClient.Conn)
}

func (m *MockRedisClient) Do(ctx context.Context, args ...interface{}) *redisClient.Cmd {
	mockArgs := m.Called(append([]interface{}{ctx}, args...)...)
	return mockArgs.Get(0).(*redisClient.Cmd)
}
