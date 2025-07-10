package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type JobCategory string

const (
	JobCategoryMowing      JobCategory = "mowing"
	JobCategoryWeeding     JobCategory = "weeding"
	JobCategoryLeafRemoval JobCategory = "leaf_removal"
	JobCategoryTrimming    JobCategory = "trimming"
	JobCategoryCleanup     JobCategory = "cleanup"
	JobCategoryOther       JobCategory = "other"
)

type JobStatus string

const (
	JobStatusOpen        JobStatus = "open"
	JobStatusInProgress  JobStatus = "in_progress"
	JobStatusCompleted   JobStatus = "completed"
	JobStatusCancelled   JobStatus = "cancelled"
)

type Visibility string

const (
	VisibilityZipCode       Visibility = "zip_code"
	VisibilitySchoolDistrict Visibility = "school_district"
)

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, s)
}

type Job struct {
	ID                           uint         `json:"id" gorm:"primaryKey"`
	UserID                       uint         `json:"user_id" gorm:"not null;index"`
	Title                        string       `json:"title" gorm:"not null"`
	Description                  string       `json:"description"`
	SpecialNotes                 string       `json:"special_notes"`
	Category                     JobCategory  `json:"category" gorm:"not null"`
	FixedPrice                   float64      `json:"fixed_price" gorm:"type:decimal(10,2)"`
	EstimatedHours               float64      `json:"estimated_hours" gorm:"type:decimal(4,2)"`
	Address                      string       `json:"address"`
	Latitude                     *float64     `json:"latitude"`
	Longitude                    *float64     `json:"longitude"`
	ZipCode                      string       `json:"zip_code" gorm:"index"`
	ElementarySchoolDistrictName string       `json:"elementary_school_district_name" gorm:"index"`
	Visibility                   Visibility   `json:"visibility" gorm:"not null"`
	Status                       JobStatus    `json:"status" gorm:"default:open;index"`
	ScheduledDate                *time.Time   `json:"scheduled_date"`
	CreatedAt                    time.Time    `json:"created_at"`
	UpdatedAt                    time.Time    `json:"updated_at"`
	CompletionImageUrls          StringArray  `json:"completion_image_urls" gorm:"type:jsonb"`

	// Relationships
	User         User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Applications []JobApplication `json:"applications,omitempty" gorm:"foreignKey:JobID"`
	Reviews      []Review         `json:"reviews,omitempty" gorm:"foreignKey:JobID"`
	Payments     []Payment        `json:"payments,omitempty" gorm:"foreignKey:RelatedID;where:type = 'job_payment'"`
}

func (j *Job) BeforeCreate(tx *gorm.DB) error {
	j.CreatedAt = time.Now()
	j.UpdatedAt = time.Now()
	return nil
}

func (j *Job) BeforeUpdate(tx *gorm.DB) error {
	j.UpdatedAt = time.Now()
	return nil
}

type JobResponse struct {
	ID                           uint        `json:"id"`
	Title                        string      `json:"title"`
	Description                  string      `json:"description"`
	SpecialNotes                 string      `json:"special_notes"`
	Category                     JobCategory `json:"category"`
	FixedPrice                   float64     `json:"fixed_price"`
	EstimatedHours               float64     `json:"estimated_hours"`
	Address                      string      `json:"address"`
	ZipCode                      string      `json:"zip_code"`
	ElementarySchoolDistrictName string      `json:"elementary_school_district_name"`
	Visibility                   Visibility  `json:"visibility"`
	Status                       JobStatus   `json:"status"`
	ScheduledDate                *time.Time  `json:"scheduled_date"`
	CreatedAt                    time.Time   `json:"created_at"`
	UpdatedAt                    time.Time   `json:"updated_at"`
	CompletionImageUrls          StringArray `json:"completion_image_urls"`
	User                         UserPublicProfile `json:"user"`
}

func (j *Job) ToResponse() JobResponse {
	return JobResponse{
		ID:                           j.ID,
		Title:                        j.Title,
		Description:                  j.Description,
		SpecialNotes:                 j.SpecialNotes,
		Category:                     j.Category,
		FixedPrice:                   j.FixedPrice,
		EstimatedHours:               j.EstimatedHours,
		Address:                      j.Address,
		ZipCode:                      j.ZipCode,
		ElementarySchoolDistrictName: j.ElementarySchoolDistrictName,
		Visibility:                   j.Visibility,
		Status:                       j.Status,
		ScheduledDate:                j.ScheduledDate,
		CreatedAt:                    j.CreatedAt,
		UpdatedAt:                    j.UpdatedAt,
		CompletionImageUrls:          j.CompletionImageUrls,
		User:                         j.User.ToPublicProfile(),
	}
}