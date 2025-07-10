package models

import (
	"time"

	"gorm.io/gorm"
)

type RentalStatus string

const (
	RentalStatusRequested RentalStatus = "requested"
	RentalStatusApproved  RentalStatus = "approved"
	RentalStatusActive    RentalStatus = "active"
	RentalStatusCompleted RentalStatus = "completed"
	RentalStatusCancelled RentalStatus = "cancelled"
)

type EquipmentRental struct {
	ID           uint         `json:"id" gorm:"primaryKey"`
	EquipmentID  uint         `json:"equipment_id" gorm:"not null;index"`
	RenterUserID uint         `json:"renter_user_id" gorm:"not null;index"`
	StartDate    time.Time    `json:"start_date" gorm:"not null"`
	EndDate      time.Time    `json:"end_date" gorm:"not null"`
	TotalPrice   float64      `json:"total_price" gorm:"type:decimal(10,2)"`
	Status       RentalStatus `json:"status" gorm:"default:requested"`
	PickupNotes  string       `json:"pickup_notes"`
	ReturnNotes  string       `json:"return_notes"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`

	// Relationships
	Equipment Equipment `json:"equipment,omitempty" gorm:"foreignKey:EquipmentID"`
	Renter    User      `json:"renter,omitempty" gorm:"foreignKey:RenterUserID"`
	Reviews   []Review  `json:"reviews,omitempty" gorm:"foreignKey:EquipmentRentalID"`
	Payments  []Payment `json:"payments,omitempty" gorm:"foreignKey:RelatedID;where:type = 'equipment_rental'"`
}

func (er *EquipmentRental) BeforeCreate(tx *gorm.DB) error {
	er.CreatedAt = time.Now()
	er.UpdatedAt = time.Now()
	return nil
}

func (er *EquipmentRental) BeforeUpdate(tx *gorm.DB) error {
	er.UpdatedAt = time.Now()
	return nil
}

type EquipmentRentalResponse struct {
	ID           uint              `json:"id"`
	StartDate    time.Time         `json:"start_date"`
	EndDate      time.Time         `json:"end_date"`
	TotalPrice   float64           `json:"total_price"`
	Status       RentalStatus      `json:"status"`
	PickupNotes  string            `json:"pickup_notes"`
	ReturnNotes  string            `json:"return_notes"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Equipment    EquipmentResponse `json:"equipment"`
	Renter       UserPublicProfile `json:"renter"`
}

func (er *EquipmentRental) ToResponse() EquipmentRentalResponse {
	return EquipmentRentalResponse{
		ID:          er.ID,
		StartDate:   er.StartDate,
		EndDate:     er.EndDate,
		TotalPrice:  er.TotalPrice,
		Status:      er.Status,
		PickupNotes: er.PickupNotes,
		ReturnNotes: er.ReturnNotes,
		CreatedAt:   er.CreatedAt,
		UpdatedAt:   er.UpdatedAt,
		Equipment:   er.Equipment.ToResponse(),
		Renter:      er.Renter.ToPublicProfile(),
	}
}