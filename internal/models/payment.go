package models

import (
	"time"

	"gorm.io/gorm"
)

type PaymentType string

const (
	PaymentTypeJobPayment        PaymentType = "job_payment"
	PaymentTypeEquipmentRental   PaymentType = "equipment_rental"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

type Payment struct {
	ID                    uint          `json:"id" gorm:"primaryKey"`
	UserID                uint          `json:"user_id" gorm:"not null;index"`
	StripePaymentIntentID string        `json:"stripe_payment_intent_id" gorm:"not null"`
	Amount                float64       `json:"amount" gorm:"type:decimal(10,2)"`
	Currency              string        `json:"currency" gorm:"default:usd"`
	Type                  PaymentType   `json:"type" gorm:"not null"`
	RelatedID             uint          `json:"related_id" gorm:"not null;index"`
	Status                PaymentStatus `json:"status" gorm:"default:pending"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Payment) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

type PaymentResponse struct {
	ID                    uint          `json:"id"`
	StripePaymentIntentID string        `json:"stripe_payment_intent_id"`
	Amount                float64       `json:"amount"`
	Currency              string        `json:"currency"`
	Type                  PaymentType   `json:"type"`
	RelatedID             uint          `json:"related_id"`
	Status                PaymentStatus `json:"status"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}

func (p *Payment) ToResponse() PaymentResponse {
	return PaymentResponse{
		ID:                    p.ID,
		StripePaymentIntentID: p.StripePaymentIntentID,
		Amount:                p.Amount,
		Currency:              p.Currency,
		Type:                  p.Type,
		RelatedID:             p.RelatedID,
		Status:                p.Status,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}