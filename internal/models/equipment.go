package models

import (
	"time"

	"gorm.io/gorm"
)

type EquipmentCategory string

const (
	EquipmentCategoryMower         EquipmentCategory = "mower"
	EquipmentCategoryWeedWhacker   EquipmentCategory = "weed_whacker"
	EquipmentCategoryEdger         EquipmentCategory = "edger"
)

type FuelType string

const (
	FuelTypeGas      FuelType = "gas"
	FuelTypeElectric FuelType = "electric"
	FuelTypeBattery  FuelType = "battery"
)

type PowerType string

const (
	PowerTypeCorded   PowerType = "corded"
	PowerTypeCordless PowerType = "cordless"
	PowerTypeGas      PowerType = "gas"
	PowerTypePush     PowerType = "push"
)

type Equipment struct {
	ID                           uint              `json:"id" gorm:"primaryKey"`
	UserID                       uint              `json:"user_id" gorm:"not null;index"`
	Name                         string            `json:"name" gorm:"not null"`
	Make                         string            `json:"make"`
	Model                        string            `json:"model"`
	Category                     EquipmentCategory `json:"category" gorm:"not null"`
	FuelType                     FuelType          `json:"fuel_type"`
	PowerType                    PowerType         `json:"power_type"`
	DailyRentalPrice             float64           `json:"daily_rental_price" gorm:"type:decimal(10,2)"`
	Description                  string            `json:"description"`
	ImageUrls                    StringArray       `json:"image_urls" gorm:"type:jsonb"`
	IsAvailable                  bool              `json:"is_available" gorm:"default:true"`
	Address                      string            `json:"address"`
	Latitude                     *float64          `json:"latitude"`
	Longitude                    *float64          `json:"longitude"`
	ZipCode                      string            `json:"zip_code" gorm:"index"`
	ElementarySchoolDistrictName string            `json:"elementary_school_district_name" gorm:"index"`
	Visibility                   Visibility        `json:"visibility" gorm:"not null"`
	CreatedAt                    time.Time         `json:"created_at"`
	UpdatedAt                    time.Time         `json:"updated_at"`

	// Relationships
	User    User                `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Rentals []EquipmentRental   `json:"rentals,omitempty" gorm:"foreignKey:EquipmentID"`
	Reviews []Review            `json:"reviews,omitempty" gorm:"foreignKey:EquipmentRentalID"`
}

func (e *Equipment) BeforeCreate(tx *gorm.DB) error {
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Equipment) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}

type EquipmentResponse struct {
	ID                           uint              `json:"id"`
	Name                         string            `json:"name"`
	Make                         string            `json:"make"`
	Model                        string            `json:"model"`
	Category                     EquipmentCategory `json:"category"`
	FuelType                     FuelType          `json:"fuel_type"`
	PowerType                    PowerType         `json:"power_type"`
	DailyRentalPrice             float64           `json:"daily_rental_price"`
	Description                  string            `json:"description"`
	ImageUrls                    StringArray       `json:"image_urls"`
	IsAvailable                  bool              `json:"is_available"`
	Address                      string            `json:"address"`
	ZipCode                      string            `json:"zip_code"`
	ElementarySchoolDistrictName string            `json:"elementary_school_district_name"`
	Visibility                   Visibility        `json:"visibility"`
	CreatedAt                    time.Time         `json:"created_at"`
	UpdatedAt                    time.Time         `json:"updated_at"`
	User                         UserPublicProfile `json:"user"`
}

func (e *Equipment) ToResponse() EquipmentResponse {
	return EquipmentResponse{
		ID:                           e.ID,
		Name:                         e.Name,
		Make:                         e.Make,
		Model:                        e.Model,
		Category:                     e.Category,
		FuelType:                     e.FuelType,
		PowerType:                    e.PowerType,
		DailyRentalPrice:             e.DailyRentalPrice,
		Description:                  e.Description,
		ImageUrls:                    e.ImageUrls,
		IsAvailable:                  e.IsAvailable,
		Address:                      e.Address,
		ZipCode:                      e.ZipCode,
		ElementarySchoolDistrictName: e.ElementarySchoolDistrictName,
		Visibility:                   e.Visibility,
		CreatedAt:                    e.CreatedAt,
		UpdatedAt:                    e.UpdatedAt,
		User:                         e.User.ToPublicProfile(),
	}
}