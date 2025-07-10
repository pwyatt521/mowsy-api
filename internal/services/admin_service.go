package services

import (
	"errors"
	"fmt"
	"time"

	"mowsy-api/internal/models"
	"mowsy-api/pkg/database"

	"gorm.io/gorm"
)

type AdminService struct {
	db *gorm.DB
}

func NewAdminService() *AdminService {
	return &AdminService{
		db: database.GetDB(),
	}
}

type AdminUserListFilters struct {
	IsActive            *bool  `form:"is_active"`
	InsuranceVerified   *bool  `form:"insurance_verified"`
	ZipCode             string `form:"zip_code"`
	SchoolDistrict      string `form:"school_district"`
	Page                int    `form:"page"`
	Limit               int    `form:"limit"`
}

func (s *AdminService) GetUsers(filters AdminUserListFilters) ([]models.UserResponse, error) {
	query := s.db.Model(&models.User{})

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	if filters.InsuranceVerified != nil {
		query = query.Where("insurance_verified = ?", *filters.InsuranceVerified)
	}

	if filters.ZipCode != "" {
		query = query.Where("zip_code = ?", filters.ZipCode)
	}

	if filters.SchoolDistrict != "" {
		query = query.Where("elementary_school_district_name = ?", filters.SchoolDistrict)
	}

	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}

	offset := (filters.Page - 1) * filters.Limit

	var users []models.User
	if err := query.Order("created_at DESC").Offset(offset).Limit(filters.Limit).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, nil
}

func (s *AdminService) DeactivateUser(userID uint) error {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if err := s.db.Model(&user).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

func (s *AdminService) ActivateUser(userID uint) error {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if err := s.db.Model(&user).Update("is_active", true).Error; err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

func (s *AdminService) VerifyInsurance(userID uint) error {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user.InsuranceDocumentURL == "" {
		return errors.New("user has not uploaded insurance document")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"insurance_verified":    true,
		"insurance_verified_at": &now,
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to verify insurance: %w", err)
	}

	return nil
}

func (s *AdminService) RemoveJob(jobID uint) error {
	var job models.Job
	if err := s.db.Where("id = ?", jobID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("job not found")
		}
		return fmt.Errorf("failed to find job: %w", err)
	}

	if err := s.db.Delete(&job).Error; err != nil {
		return fmt.Errorf("failed to remove job: %w", err)
	}

	return nil
}

func (s *AdminService) RemoveEquipment(equipmentID uint) error {
	var equipment models.Equipment
	if err := s.db.Where("id = ?", equipmentID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("equipment not found")
		}
		return fmt.Errorf("failed to find equipment: %w", err)
	}

	var activeRentals []models.EquipmentRental
	if err := s.db.Where("equipment_id = ? AND status IN (?)", equipmentID, 
		[]models.RentalStatus{models.RentalStatusApproved, models.RentalStatusActive}).
		Find(&activeRentals).Error; err != nil {
		return fmt.Errorf("failed to check active rentals: %w", err)
	}

	if len(activeRentals) > 0 {
		return errors.New("cannot remove equipment with active rentals")
	}

	if err := s.db.Delete(&equipment).Error; err != nil {
		return fmt.Errorf("failed to remove equipment: %w", err)
	}

	return nil
}

type AdminStats struct {
	TotalUsers              int64 `json:"total_users"`
	ActiveUsers             int64 `json:"active_users"`
	VerifiedInsuranceUsers  int64 `json:"verified_insurance_users"`
	TotalJobs               int64 `json:"total_jobs"`
	OpenJobs                int64 `json:"open_jobs"`
	CompletedJobs           int64 `json:"completed_jobs"`
	TotalEquipment          int64 `json:"total_equipment"`
	AvailableEquipment      int64 `json:"available_equipment"`
	TotalRentals            int64 `json:"total_rentals"`
	ActiveRentals           int64 `json:"active_rentals"`
	CompletedRentals        int64 `json:"completed_rentals"`
	TotalPayments           int64 `json:"total_payments"`
	SuccessfulPayments      int64 `json:"successful_payments"`
}

func (s *AdminService) GetStats() (*AdminStats, error) {
	stats := &AdminStats{}

	if err := s.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	if err := s.db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	if err := s.db.Model(&models.User{}).Where("insurance_verified = ?", true).Count(&stats.VerifiedInsuranceUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count verified insurance users: %w", err)
	}

	if err := s.db.Model(&models.Job{}).Count(&stats.TotalJobs).Error; err != nil {
		return nil, fmt.Errorf("failed to count total jobs: %w", err)
	}

	if err := s.db.Model(&models.Job{}).Where("status = ?", models.JobStatusOpen).Count(&stats.OpenJobs).Error; err != nil {
		return nil, fmt.Errorf("failed to count open jobs: %w", err)
	}

	if err := s.db.Model(&models.Job{}).Where("status = ?", models.JobStatusCompleted).Count(&stats.CompletedJobs).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed jobs: %w", err)
	}

	if err := s.db.Model(&models.Equipment{}).Count(&stats.TotalEquipment).Error; err != nil {
		return nil, fmt.Errorf("failed to count total equipment: %w", err)
	}

	if err := s.db.Model(&models.Equipment{}).Where("is_available = ?", true).Count(&stats.AvailableEquipment).Error; err != nil {
		return nil, fmt.Errorf("failed to count available equipment: %w", err)
	}

	if err := s.db.Model(&models.EquipmentRental{}).Count(&stats.TotalRentals).Error; err != nil {
		return nil, fmt.Errorf("failed to count total rentals: %w", err)
	}

	if err := s.db.Model(&models.EquipmentRental{}).Where("status = ?", models.RentalStatusActive).Count(&stats.ActiveRentals).Error; err != nil {
		return nil, fmt.Errorf("failed to count active rentals: %w", err)
	}

	if err := s.db.Model(&models.EquipmentRental{}).Where("status = ?", models.RentalStatusCompleted).Count(&stats.CompletedRentals).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed rentals: %w", err)
	}

	if err := s.db.Model(&models.Payment{}).Count(&stats.TotalPayments).Error; err != nil {
		return nil, fmt.Errorf("failed to count total payments: %w", err)
	}

	if err := s.db.Model(&models.Payment{}).Where("status = ?", models.PaymentStatusSucceeded).Count(&stats.SuccessfulPayments).Error; err != nil {
		return nil, fmt.Errorf("failed to count successful payments: %w", err)
	}

	return stats, nil
}