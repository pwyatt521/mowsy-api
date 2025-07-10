package testutils

import (
	"io"
	"time"

	"mowsy-api/internal/models"
	"mowsy-api/pkg/storage"

	"github.com/stretchr/testify/mock"
)

// MockS3Service is a mock implementation of the S3Service
type MockS3Service struct {
	mock.Mock
}

func (m *MockS3Service) UploadFile(file io.Reader, fileName, mimeType string, userID uint) (*storage.UploadResult, error) {
	args := m.Called(file, fileName, mimeType, userID)
	return args.Get(0).(*storage.UploadResult), args.Error(1)
}

func (m *MockS3Service) DeleteFile(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockS3Service) GetPresignedURL(key string, expiration time.Duration) (string, error) {
	args := m.Called(key, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockS3Service) GetPresignedUploadURL(key string, mimeType string, expiration time.Duration) (string, error) {
	args := m.Called(key, mimeType, expiration)
	return args.String(0), args.Error(1)
}

// MockGeocodioService is a mock implementation of the Geocodio service
type MockGeocodioService struct {
	mock.Mock
}

func (m *MockGeocodioService) GeocodeUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockGeocodioService) GeocodeJob(job *models.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockGeocodioService) GeocodeEquipment(equipment *models.Equipment) error {
	args := m.Called(equipment)
	return args.Error(0)
}

// MockStripeService is a mock for Stripe operations
type MockStripeService struct {
	mock.Mock
}

func (m *MockStripeService) CreateCustomer(email, name, phone string) (string, error) {
	args := m.Called(email, name, phone)
	return args.String(0), args.Error(1)
}

func (m *MockStripeService) CreatePaymentIntent(amount int64, currency, customerID string, metadata map[string]string) (string, string, error) {
	args := m.Called(amount, currency, customerID, metadata)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockStripeService) GetPaymentIntent(paymentIntentID string) (string, error) {
	args := m.Called(paymentIntentID)
	return args.String(0), args.Error(1)
}

// SetEnvironmentForTesting sets up environment variables for testing
func SetEnvironmentForTesting() {
	// Set test environment variables
	// Note: In real tests, you'd want to use t.Setenv() in Go 1.17+
}

// ResetEnvironmentAfterTesting cleans up environment variables after testing
func ResetEnvironmentAfterTesting() {
	// Clean up environment variables
}