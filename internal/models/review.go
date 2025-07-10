package models

import (
	"time"

	"gorm.io/gorm"
)

type ReviewType string

const (
	ReviewTypeJobCompletion    ReviewType = "job_completion"
	ReviewTypeEquipmentRental  ReviewType = "equipment_rental"
)

type Review struct {
	ID                  uint       `json:"id" gorm:"primaryKey"`
	ReviewerUserID      uint       `json:"reviewer_user_id" gorm:"not null;index"`
	ReviewedUserID      uint       `json:"reviewed_user_id" gorm:"not null;index"`
	JobID               *uint      `json:"job_id,omitempty" gorm:"index"`
	EquipmentRentalID   *uint      `json:"equipment_rental_id,omitempty" gorm:"index"`
	Rating              int        `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment             string     `json:"comment"`
	Type                ReviewType `json:"type" gorm:"not null"`
	CreatedAt           time.Time  `json:"created_at"`

	// Relationships
	Reviewer        User             `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerUserID"`
	ReviewedUser    User             `json:"reviewed_user,omitempty" gorm:"foreignKey:ReviewedUserID"`
	Job             Job              `json:"job,omitempty" gorm:"foreignKey:JobID"`
	EquipmentRental EquipmentRental  `json:"equipment_rental,omitempty" gorm:"foreignKey:EquipmentRentalID"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	r.CreatedAt = time.Now()
	return nil
}

type ReviewResponse struct {
	ID        uint              `json:"id"`
	Rating    int               `json:"rating"`
	Comment   string            `json:"comment"`
	Type      ReviewType        `json:"type"`
	CreatedAt time.Time         `json:"created_at"`
	Reviewer  UserPublicProfile `json:"reviewer"`
}

func (r *Review) ToResponse() ReviewResponse {
	return ReviewResponse{
		ID:        r.ID,
		Rating:    r.Rating,
		Comment:   r.Comment,
		Type:      r.Type,
		CreatedAt: r.CreatedAt,
		Reviewer:  r.Reviewer.ToPublicProfile(),
	}
}