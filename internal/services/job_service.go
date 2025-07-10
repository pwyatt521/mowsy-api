package services

import (
	"errors"
	"fmt"
	"time"

	"mowsy-api/internal/models"
	"mowsy-api/internal/utils"
	"mowsy-api/pkg/database"

	"gorm.io/gorm"
)

type JobService struct {
	db *gorm.DB
}

func NewJobService() *JobService {
	return &JobService{
		db: database.GetDB(),
	}
}

type CreateJobRequest struct {
	Title            string                `json:"title" binding:"required"`
	Description      string                `json:"description"`
	SpecialNotes     string                `json:"special_notes"`
	Category         models.JobCategory    `json:"category" binding:"required"`
	FixedPrice       float64               `json:"fixed_price" binding:"required"`
	EstimatedHours   float64               `json:"estimated_hours"`
	Address          string                `json:"address"`
	Visibility       models.Visibility     `json:"visibility" binding:"required"`
	ScheduledDate    *time.Time            `json:"scheduled_date"`
}

type UpdateJobRequest struct {
	Title            string                `json:"title"`
	Description      string                `json:"description"`
	SpecialNotes     string                `json:"special_notes"`
	Category         models.JobCategory    `json:"category"`
	FixedPrice       *float64              `json:"fixed_price"`
	EstimatedHours   *float64              `json:"estimated_hours"`
	Address          string                `json:"address"`
	Visibility       models.Visibility     `json:"visibility"`
	ScheduledDate    *time.Time            `json:"scheduled_date"`
}

type JobFilters struct {
	Visibility   models.Visibility     `form:"visibility"`
	ZipCode      string                `form:"zip_code"`
	District     string                `form:"district"`
	Category     models.JobCategory    `form:"category"`
	Status       models.JobStatus      `form:"status"`
	MinPrice     *float64              `form:"min_price"`
	MaxPrice     *float64              `form:"max_price"`
	Page         int                   `form:"page"`
	Limit        int                   `form:"limit"`
}

func (s *JobService) CreateJob(userID uint, req CreateJobRequest) (*models.JobResponse, error) {
	userService := &UserService{db: s.db}
	user, err := userService.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if req.FixedPrice <= 0 {
		return nil, errors.New("fixed price must be greater than 0")
	}

	if req.EstimatedHours < 0 {
		return nil, errors.New("estimated hours cannot be negative")
	}

	job := models.Job{
		UserID:         userID,
		Title:          utils.SanitizeString(req.Title),
		Description:    utils.SanitizeString(req.Description),
		SpecialNotes:   utils.SanitizeString(req.SpecialNotes),
		Category:       req.Category,
		FixedPrice:     req.FixedPrice,
		EstimatedHours: req.EstimatedHours,
		Address:        utils.SanitizeString(req.Address),
		Visibility:     req.Visibility,
		ScheduledDate:  req.ScheduledDate,
		Status:         models.JobStatusOpen,
	}

	if job.Address != "" {
		geocodioService := NewGeocodioService()
		if err := geocodioService.GeocodeJob(&job); err != nil {
			fmt.Printf("Warning: Failed to geocode job address: %v\n", err)
		}
	} else {
		job.Latitude = user.Latitude
		job.Longitude = user.Longitude
		job.ZipCode = user.ZipCode
		job.ElementarySchoolDistrictName = user.ElementarySchoolDistrictName
	}

	if err := s.db.Create(&job).Error; err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	if err := s.db.Preload("User").First(&job, job.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load job with user: %w", err)
	}

	response := job.ToResponse()
	return &response, nil
}

func (s *JobService) GetJobs(filters JobFilters) ([]models.JobResponse, error) {
	query := s.db.Preload("User").Where("status = ?", models.JobStatusOpen)

	if filters.Visibility != "" {
		query = query.Where("visibility = ?", filters.Visibility)
	}

	if filters.ZipCode != "" {
		query = query.Where("zip_code = ?", filters.ZipCode)
	}

	if filters.District != "" {
		query = query.Where("elementary_school_district_name = ?", filters.District)
	}

	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.MinPrice != nil {
		query = query.Where("fixed_price >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("fixed_price <= ?", *filters.MaxPrice)
	}

	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}

	offset := (filters.Page - 1) * filters.Limit

	var jobs []models.Job
	if err := query.Order("created_at DESC").Offset(offset).Limit(filters.Limit).Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %w", err)
	}

	responses := make([]models.JobResponse, len(jobs))
	for i, job := range jobs {
		responses[i] = job.ToResponse()
	}

	return responses, nil
}

func (s *JobService) GetJobByID(jobID uint) (*models.JobResponse, error) {
	var job models.Job
	if err := s.db.Preload("User").Where("id = ?", jobID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found")
		}
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	response := job.ToResponse()
	return &response, nil
}

func (s *JobService) UpdateJob(jobID, userID uint, req UpdateJobRequest) (*models.JobResponse, error) {
	var job models.Job
	if err := s.db.Where("id = ? AND user_id = ?", jobID, userID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found or you don't have permission to update it")
		}
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	if job.Status != models.JobStatusOpen {
		return nil, errors.New("cannot update job that is not open")
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = utils.SanitizeString(req.Title)
	}
	if req.Description != "" {
		updates["description"] = utils.SanitizeString(req.Description)
	}
	if req.SpecialNotes != "" {
		updates["special_notes"] = utils.SanitizeString(req.SpecialNotes)
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.FixedPrice != nil {
		if *req.FixedPrice <= 0 {
			return nil, errors.New("fixed price must be greater than 0")
		}
		updates["fixed_price"] = *req.FixedPrice
	}
	if req.EstimatedHours != nil {
		if *req.EstimatedHours < 0 {
			return nil, errors.New("estimated hours cannot be negative")
		}
		updates["estimated_hours"] = *req.EstimatedHours
	}
	if req.Address != "" {
		updates["address"] = utils.SanitizeString(req.Address)
	}
	if req.Visibility != "" {
		updates["visibility"] = req.Visibility
	}
	if req.ScheduledDate != nil {
		updates["scheduled_date"] = req.ScheduledDate
	}

	if err := s.db.Model(&job).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	if req.Address != "" {
		geocodioService := NewGeocodioService()
		if err := geocodioService.GeocodeJob(&job); err != nil {
			fmt.Printf("Warning: Failed to geocode job address: %v\n", err)
		} else {
			s.db.Save(&job)
		}
	}

	if err := s.db.Preload("User").First(&job, job.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load updated job: %w", err)
	}

	response := job.ToResponse()
	return &response, nil
}

func (s *JobService) DeleteJob(jobID, userID uint) error {
	var job models.Job
	if err := s.db.Where("id = ? AND user_id = ?", jobID, userID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("job not found or you don't have permission to delete it")
		}
		return fmt.Errorf("failed to fetch job: %w", err)
	}

	if job.Status != models.JobStatusOpen {
		return errors.New("cannot delete job that is not open")
	}

	if err := s.db.Delete(&job).Error; err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

func (s *JobService) ApplyForJob(jobID, userID uint, message string) (*models.JobApplicationResponse, error) {
	var job models.Job
	if err := s.db.Where("id = ? AND status = ?", jobID, models.JobStatusOpen).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found or not accepting applications")
		}
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	if job.UserID == userID {
		return nil, errors.New("cannot apply for your own job")
	}

	var existingApplication models.JobApplication
	if err := s.db.Where("job_id = ? AND user_id = ?", jobID, userID).First(&existingApplication).Error; err == nil {
		return nil, errors.New("you have already applied for this job")
	}

	application := models.JobApplication{
		JobID:   jobID,
		UserID:  userID,
		Message: utils.SanitizeString(message),
		Status:  models.ApplicationStatusPending,
	}

	if err := s.db.Create(&application).Error; err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	if err := s.db.Preload("User").First(&application, application.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load application with user: %w", err)
	}

	response := application.ToResponse()
	return &response, nil
}

func (s *JobService) GetJobApplications(jobID, userID uint) ([]models.JobApplicationResponse, error) {
	var job models.Job
	if err := s.db.Where("id = ? AND user_id = ?", jobID, userID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found or you don't have permission to view applications")
		}
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	var applications []models.JobApplication
	if err := s.db.Where("job_id = ?", jobID).
		Preload("User").
		Order("applied_at DESC").
		Find(&applications).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch applications: %w", err)
	}

	responses := make([]models.JobApplicationResponse, len(applications))
	for i, app := range applications {
		responses[i] = app.ToResponse()
	}

	return responses, nil
}

func (s *JobService) UpdateApplicationStatus(jobID, applicationID, userID uint, status models.ApplicationStatus) error {
	var job models.Job
	if err := s.db.Where("id = ? AND user_id = ?", jobID, userID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("job not found or you don't have permission to update applications")
		}
		return fmt.Errorf("failed to fetch job: %w", err)
	}

	var application models.JobApplication
	if err := s.db.Where("id = ? AND job_id = ?", applicationID, jobID).First(&application).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("application not found")
		}
		return fmt.Errorf("failed to fetch application: %w", err)
	}

	if err := s.db.Model(&application).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update application status: %w", err)
	}

	if status == models.ApplicationStatusAccepted {
		if err := s.db.Model(&job).Update("status", models.JobStatusInProgress).Error; err != nil {
			return fmt.Errorf("failed to update job status: %w", err)
		}
	}

	return nil
}

func (s *JobService) CompleteJob(jobID, userID uint, imageUrls []string) error {
	var job models.Job
	if err := s.db.Where("id = ? AND user_id = ?", jobID, userID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("job not found or you don't have permission to complete it")
		}
		return fmt.Errorf("failed to fetch job: %w", err)
	}

	if job.Status != models.JobStatusInProgress {
		return errors.New("job must be in progress to complete")
	}

	updates := map[string]interface{}{
		"status":                 models.JobStatusCompleted,
		"completion_image_urls":  models.StringArray(imageUrls),
	}

	if err := s.db.Model(&job).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	return nil
}