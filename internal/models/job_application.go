package models

import (
	"time"

	"gorm.io/gorm"
)

type ApplicationStatus string

const (
	ApplicationStatusPending  ApplicationStatus = "pending"
	ApplicationStatusAccepted ApplicationStatus = "accepted"
	ApplicationStatusRejected ApplicationStatus = "rejected"
)

type JobApplication struct {
	ID        uint              `json:"id" gorm:"primaryKey"`
	JobID     uint              `json:"job_id" gorm:"not null;index"`
	UserID    uint              `json:"user_id" gorm:"not null;index"`
	Message   string            `json:"message"`
	AppliedAt time.Time         `json:"applied_at"`
	Status    ApplicationStatus `json:"status" gorm:"default:pending"`

	// Relationships
	Job  Job  `json:"job,omitempty" gorm:"foreignKey:JobID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (ja *JobApplication) BeforeCreate(tx *gorm.DB) error {
	ja.AppliedAt = time.Now()
	return nil
}

type JobApplicationResponse struct {
	ID        uint              `json:"id"`
	JobID     uint              `json:"job_id"`
	Message   string            `json:"message"`
	AppliedAt time.Time         `json:"applied_at"`
	Status    ApplicationStatus `json:"status"`
	User      UserPublicProfile `json:"user"`
}

func (ja *JobApplication) ToResponse() JobApplicationResponse {
	return JobApplicationResponse{
		ID:        ja.ID,
		JobID:     ja.JobID,
		Message:   ja.Message,
		AppliedAt: ja.AppliedAt,
		Status:    ja.Status,
		User:      ja.User.ToPublicProfile(),
	}
}