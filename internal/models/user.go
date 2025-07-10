package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                           uint      `json:"id" gorm:"primaryKey"`
	Email                        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash                 string    `json:"-" gorm:"not null"`
	FirstName                    string    `json:"first_name" gorm:"not null"`
	LastName                     string    `json:"last_name" gorm:"not null"`
	Phone                        string    `json:"phone"`
	Address                      string    `json:"address"`
	City                         string    `json:"city"`
	State                        string    `json:"state"`
	ZipCode                      string    `json:"zip_code" gorm:"index"`
	Latitude                     *float64  `json:"latitude"`
	Longitude                    *float64  `json:"longitude"`
	ElementarySchoolDistrictName string    `json:"elementary_school_district_name" gorm:"index"`
	ElementarySchoolDistrictCode string    `json:"elementary_school_district_code"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
	IsActive                     bool      `json:"is_active"`
	StripeCustomerID             string    `json:"stripe_customer_id"`
	InsuranceDocumentURL         string    `json:"insurance_document_url"`
	InsuranceVerified            bool      `json:"insurance_verified" gorm:"default:false"`
	InsuranceVerifiedAt          *time.Time `json:"insurance_verified_at"`

	// Relationships
	PostedJobs         []Job              `json:"posted_jobs,omitempty" gorm:"foreignKey:UserID"`
	JobApplications    []JobApplication   `json:"job_applications,omitempty" gorm:"foreignKey:UserID"`
	Equipment          []Equipment        `json:"equipment,omitempty" gorm:"foreignKey:UserID"`
	EquipmentRentals   []EquipmentRental  `json:"equipment_rentals,omitempty" gorm:"foreignKey:RenterUserID"`
	ReviewsGiven       []Review           `json:"reviews_given,omitempty" gorm:"foreignKey:ReviewerUserID"`
	ReviewsReceived    []Review           `json:"reviews_received,omitempty" gorm:"foreignKey:ReviewedUserID"`
	Payments           []Payment          `json:"payments,omitempty" gorm:"foreignKey:UserID"`
}

type UserResponse struct {
	ID                           uint      `json:"id"`
	Email                        string    `json:"email"`
	FirstName                    string    `json:"first_name"`
	LastName                     string    `json:"last_name"`
	Phone                        string    `json:"phone"`
	Address                      string    `json:"address"`
	City                         string    `json:"city"`
	State                        string    `json:"state"`
	ZipCode                      string    `json:"zip_code"`
	ElementarySchoolDistrictName string    `json:"elementary_school_district_name"`
	CreatedAt                    time.Time `json:"created_at"`
	InsuranceVerified            bool      `json:"insurance_verified"`
	InsuranceVerifiedAt          *time.Time `json:"insurance_verified_at"`
}

type UserPublicProfile struct {
	ID                           uint      `json:"id"`
	FirstName                    string    `json:"first_name"`
	LastName                     string    `json:"last_name"`
	ElementarySchoolDistrictName string    `json:"elementary_school_district_name"`
	CreatedAt                    time.Time `json:"created_at"`
	InsuranceVerified            bool      `json:"insurance_verified"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:                           u.ID,
		Email:                        u.Email,
		FirstName:                    u.FirstName,
		LastName:                     u.LastName,
		Phone:                        u.Phone,
		Address:                      u.Address,
		City:                         u.City,
		State:                        u.State,
		ZipCode:                      u.ZipCode,
		ElementarySchoolDistrictName: u.ElementarySchoolDistrictName,
		CreatedAt:                    u.CreatedAt,
		InsuranceVerified:            u.InsuranceVerified,
		InsuranceVerifiedAt:          u.InsuranceVerifiedAt,
	}
}

func (u *User) ToPublicProfile() UserPublicProfile {
	return UserPublicProfile{
		ID:                           u.ID,
		FirstName:                    u.FirstName,
		LastName:                     u.LastName,
		ElementarySchoolDistrictName: u.ElementarySchoolDistrictName,
		CreatedAt:                    u.CreatedAt,
		InsuranceVerified:            u.InsuranceVerified,
	}
}