package services

import (
	"errors"
	"fmt"
	"os"

	"mowsy-api/internal/models"
	"mowsy-api/pkg/database"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/customer"
	"github.com/stripe/stripe-go/v75/paymentintent"
	"gorm.io/gorm"
)

type PaymentService struct {
	db *gorm.DB
}

func NewPaymentService() *PaymentService {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return &PaymentService{
		db: database.GetDB(),
	}
}

type CreatePaymentIntentRequest struct {
	Amount      float64             `json:"amount" binding:"required"`
	Currency    string              `json:"currency"`
	Type        models.PaymentType  `json:"type" binding:"required"`
	RelatedID   uint                `json:"related_id" binding:"required"`
	Description string              `json:"description"`
}

type PaymentIntentResponse struct {
	ClientSecret string `json:"client_secret"`
	PaymentID    uint   `json:"payment_id"`
}

func (s *PaymentService) CreatePaymentIntent(userID uint, req CreatePaymentIntentRequest) (*PaymentIntentResponse, error) {
	user, err := NewUserService().GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if req.Currency == "" {
		req.Currency = "usd"
	}

	if err := s.validatePaymentContext(req.Type, req.RelatedID, userID); err != nil {
		return nil, err
	}

	if user.StripeCustomerID == "" {
		customerParams := &stripe.CustomerParams{
			Email: stripe.String(user.Email),
			Name:  stripe.String(user.FirstName + " " + user.LastName),
		}
		if user.Phone != "" {
			customerParams.Phone = stripe.String(user.Phone)
		}

		stripeCustomer, err := customer.New(customerParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
		}

		if err := s.db.Model(user).Update("stripe_customer_id", stripeCustomer.ID).Error; err != nil {
			return nil, fmt.Errorf("failed to update user with Stripe customer ID: %w", err)
		}
		user.StripeCustomerID = stripeCustomer.ID
	}

	amountCents := int64(req.Amount * 100)
	paymentIntentParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(req.Currency),
		Customer: stripe.String(user.StripeCustomerID),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	if req.Description != "" {
		paymentIntentParams.Description = stripe.String(req.Description)
	}

	paymentIntentParams.Metadata = map[string]string{
		"user_id":    fmt.Sprintf("%d", userID),
		"type":       string(req.Type),
		"related_id": fmt.Sprintf("%d", req.RelatedID),
	}

	stripePaymentIntent, err := paymentintent.New(paymentIntentParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe payment intent: %w", err)
	}

	payment := models.Payment{
		UserID:                userID,
		StripePaymentIntentID: stripePaymentIntent.ID,
		Amount:                req.Amount,
		Currency:              req.Currency,
		Type:                  req.Type,
		RelatedID:             req.RelatedID,
		Status:                models.PaymentStatusPending,
	}

	if err := s.db.Create(&payment).Error; err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	return &PaymentIntentResponse{
		ClientSecret: stripePaymentIntent.ClientSecret,
		PaymentID:    payment.ID,
	}, nil
}

func (s *PaymentService) validatePaymentContext(paymentType models.PaymentType, relatedID, userID uint) error {
	switch paymentType {
	case models.PaymentTypeJobPayment:
		var job models.Job
		if err := s.db.Where("id = ? AND user_id = ? AND status = ?", relatedID, userID, models.JobStatusCompleted).First(&job).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("job not found, not owned by user, or not completed")
			}
			return fmt.Errorf("failed to validate job: %w", err)
		}

	case models.PaymentTypeEquipmentRental:
		var rental models.EquipmentRental
		if err := s.db.Preload("Equipment").Where("id = ? AND renter_user_id = ? AND status = ?", relatedID, userID, models.RentalStatusApproved).First(&rental).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("rental not found, not owned by user, or not approved")
			}
			return fmt.Errorf("failed to validate rental: %w", err)
		}

	default:
		return errors.New("invalid payment type")
	}

	return nil
}

func (s *PaymentService) ConfirmPayment(paymentID uint, userID uint) (*models.PaymentResponse, error) {
	var payment models.Payment
	if err := s.db.Where("id = ? AND user_id = ?", paymentID, userID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	stripePaymentIntent, err := paymentintent.Get(payment.StripePaymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe payment intent: %w", err)
	}

	var status models.PaymentStatus
	switch stripePaymentIntent.Status {
	case stripe.PaymentIntentStatusSucceeded:
		status = models.PaymentStatusSucceeded
	case stripe.PaymentIntentStatusCanceled:
		status = models.PaymentStatusCancelled
	case stripe.PaymentIntentStatusProcessing, stripe.PaymentIntentStatusRequiresPaymentMethod, stripe.PaymentIntentStatusRequiresConfirmation:
		status = models.PaymentStatusPending
	default:
		status = models.PaymentStatusFailed
	}

	if err := s.db.Model(&payment).Update("status", status).Error; err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	if status == models.PaymentStatusSucceeded {
		if err := s.handleSuccessfulPayment(&payment); err != nil {
			fmt.Printf("Warning: Failed to handle successful payment: %v\n", err)
		}
	}

	payment.Status = status
	response := payment.ToResponse()
	return &response, nil
}

func (s *PaymentService) handleSuccessfulPayment(payment *models.Payment) error {
	switch payment.Type {
	case models.PaymentTypeJobPayment:
		// Job payment is handled when job is marked as completed
		return nil

	case models.PaymentTypeEquipmentRental:
		var rental models.EquipmentRental
		if err := s.db.Where("id = ?", payment.RelatedID).First(&rental).Error; err != nil {
			return fmt.Errorf("failed to fetch rental: %w", err)
		}

		if rental.Status == models.RentalStatusApproved {
			if err := s.db.Model(&rental).Update("status", models.RentalStatusActive).Error; err != nil {
				return fmt.Errorf("failed to update rental status: %w", err)
			}
		}

	default:
		return errors.New("unknown payment type")
	}

	return nil
}

func (s *PaymentService) GetPaymentHistory(userID uint, page, limit int) ([]models.PaymentResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var payments []models.Payment
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch payment history: %w", err)
	}

	responses := make([]models.PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = payment.ToResponse()
	}

	return responses, nil
}

func (s *PaymentService) GetPaymentByID(paymentID, userID uint) (*models.PaymentResponse, error) {
	var payment models.Payment
	if err := s.db.Where("id = ? AND user_id = ?", paymentID, userID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	response := payment.ToResponse()
	return &response, nil
}