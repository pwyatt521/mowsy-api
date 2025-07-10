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

type EquipmentService struct {
	db *gorm.DB
}

func NewEquipmentService() *EquipmentService {
	return &EquipmentService{
		db: database.GetDB(),
	}
}

type CreateEquipmentRequest struct {
	Name             string                    `json:"name" binding:"required"`
	Make             string                    `json:"make"`
	Model            string                    `json:"model"`
	Category         models.EquipmentCategory  `json:"category" binding:"required"`
	FuelType         models.FuelType           `json:"fuel_type"`
	PowerType        models.PowerType          `json:"power_type"`
	DailyRentalPrice float64                   `json:"daily_rental_price" binding:"required"`
	Description      string                    `json:"description"`
	ImageUrls        []string                  `json:"image_urls"`
	Address          string                    `json:"address"`
	Visibility       models.Visibility         `json:"visibility" binding:"required"`
}

type UpdateEquipmentRequest struct {
	Name             string                    `json:"name"`
	Make             string                    `json:"make"`
	Model            string                    `json:"model"`
	Category         models.EquipmentCategory  `json:"category"`
	FuelType         models.FuelType           `json:"fuel_type"`
	PowerType        models.PowerType          `json:"power_type"`
	DailyRentalPrice *float64                  `json:"daily_rental_price"`
	Description      string                    `json:"description"`
	ImageUrls        []string                  `json:"image_urls"`
	Address          string                    `json:"address"`
	Visibility       models.Visibility         `json:"visibility"`
	IsAvailable      *bool                     `json:"is_available"`
}

type EquipmentFilters struct {
	Visibility   models.Visibility        `form:"visibility"`
	ZipCode      string                   `form:"zip_code"`
	District     string                   `form:"district"`
	Category     models.EquipmentCategory `form:"category"`
	FuelType     models.FuelType          `form:"fuel_type"`
	PowerType    models.PowerType         `form:"power_type"`
	MinPrice     *float64                 `form:"min_price"`
	MaxPrice     *float64                 `form:"max_price"`
	IsAvailable  *bool                    `form:"is_available"`
	Page         int                      `form:"page"`
	Limit        int                      `form:"limit"`
}

func (s *EquipmentService) CreateEquipment(userID uint, req CreateEquipmentRequest) (*models.EquipmentResponse, error) {
	user, err := NewUserService().GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if req.DailyRentalPrice <= 0 {
		return nil, errors.New("daily rental price must be greater than 0")
	}

	equipment := models.Equipment{
		UserID:           userID,
		Name:             utils.SanitizeString(req.Name),
		Make:             utils.SanitizeString(req.Make),
		Model:            utils.SanitizeString(req.Model),
		Category:         req.Category,
		FuelType:         req.FuelType,
		PowerType:        req.PowerType,
		DailyRentalPrice: req.DailyRentalPrice,
		Description:      utils.SanitizeString(req.Description),
		ImageUrls:        models.StringArray(req.ImageUrls),
		Address:          utils.SanitizeString(req.Address),
		Visibility:       req.Visibility,
		IsAvailable:      true,
	}

	if equipment.Address != "" {
		geocodioService := NewGeocodioService()
		if err := geocodioService.GeocodeEquipment(&equipment); err != nil {
			fmt.Printf("Warning: Failed to geocode equipment address: %v\n", err)
		}
	} else {
		equipment.Latitude = user.Latitude
		equipment.Longitude = user.Longitude
		equipment.ZipCode = user.ZipCode
		equipment.ElementarySchoolDistrictName = user.ElementarySchoolDistrictName
	}

	if err := s.db.Create(&equipment).Error; err != nil {
		return nil, fmt.Errorf("failed to create equipment: %w", err)
	}

	if err := s.db.Preload("User").First(&equipment, equipment.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load equipment with user: %w", err)
	}

	response := equipment.ToResponse()
	return &response, nil
}

func (s *EquipmentService) GetEquipment(filters EquipmentFilters) ([]models.EquipmentResponse, error) {
	query := s.db.Preload("User")

	if filters.IsAvailable == nil {
		available := true
		filters.IsAvailable = &available
	}
	query = query.Where("is_available = ?", *filters.IsAvailable)

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

	if filters.FuelType != "" {
		query = query.Where("fuel_type = ?", filters.FuelType)
	}

	if filters.PowerType != "" {
		query = query.Where("power_type = ?", filters.PowerType)
	}

	if filters.MinPrice != nil {
		query = query.Where("daily_rental_price >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("daily_rental_price <= ?", *filters.MaxPrice)
	}

	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}

	offset := (filters.Page - 1) * filters.Limit

	var equipment []models.Equipment
	if err := query.Order("created_at DESC").Offset(offset).Limit(filters.Limit).Find(&equipment).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch equipment: %w", err)
	}

	responses := make([]models.EquipmentResponse, len(equipment))
	for i, eq := range equipment {
		responses[i] = eq.ToResponse()
	}

	return responses, nil
}

func (s *EquipmentService) GetEquipmentByID(equipmentID uint) (*models.EquipmentResponse, error) {
	var equipment models.Equipment
	if err := s.db.Preload("User").Where("id = ?", equipmentID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("equipment not found")
		}
		return nil, fmt.Errorf("failed to fetch equipment: %w", err)
	}

	response := equipment.ToResponse()
	return &response, nil
}

func (s *EquipmentService) UpdateEquipment(equipmentID, userID uint, req UpdateEquipmentRequest) (*models.EquipmentResponse, error) {
	var equipment models.Equipment
	if err := s.db.Where("id = ? AND user_id = ?", equipmentID, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("equipment not found or you don't have permission to update it")
		}
		return nil, fmt.Errorf("failed to fetch equipment: %w", err)
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = utils.SanitizeString(req.Name)
	}
	if req.Make != "" {
		updates["make"] = utils.SanitizeString(req.Make)
	}
	if req.Model != "" {
		updates["model"] = utils.SanitizeString(req.Model)
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.FuelType != "" {
		updates["fuel_type"] = req.FuelType
	}
	if req.PowerType != "" {
		updates["power_type"] = req.PowerType
	}
	if req.DailyRentalPrice != nil {
		if *req.DailyRentalPrice <= 0 {
			return nil, errors.New("daily rental price must be greater than 0")
		}
		updates["daily_rental_price"] = *req.DailyRentalPrice
	}
	if req.Description != "" {
		updates["description"] = utils.SanitizeString(req.Description)
	}
	if req.ImageUrls != nil {
		updates["image_urls"] = models.StringArray(req.ImageUrls)
	}
	if req.Address != "" {
		updates["address"] = utils.SanitizeString(req.Address)
	}
	if req.Visibility != "" {
		updates["visibility"] = req.Visibility
	}
	if req.IsAvailable != nil {
		updates["is_available"] = *req.IsAvailable
	}

	if err := s.db.Model(&equipment).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update equipment: %w", err)
	}

	if req.Address != "" {
		geocodioService := NewGeocodioService()
		if err := geocodioService.GeocodeEquipment(&equipment); err != nil {
			fmt.Printf("Warning: Failed to geocode equipment address: %v\n", err)
		} else {
			s.db.Save(&equipment)
		}
	}

	if err := s.db.Preload("User").First(&equipment, equipment.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load updated equipment: %w", err)
	}

	response := equipment.ToResponse()
	return &response, nil
}

func (s *EquipmentService) DeleteEquipment(equipmentID, userID uint) error {
	var equipment models.Equipment
	if err := s.db.Where("id = ? AND user_id = ?", equipmentID, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("equipment not found or you don't have permission to delete it")
		}
		return fmt.Errorf("failed to fetch equipment: %w", err)
	}

	var activeRentals []models.EquipmentRental
	if err := s.db.Where("equipment_id = ? AND status IN (?)", equipmentID, 
		[]models.RentalStatus{models.RentalStatusApproved, models.RentalStatusActive}).
		Find(&activeRentals).Error; err != nil {
		return fmt.Errorf("failed to check active rentals: %w", err)
	}

	if len(activeRentals) > 0 {
		return errors.New("cannot delete equipment with active rentals")
	}

	if err := s.db.Delete(&equipment).Error; err != nil {
		return fmt.Errorf("failed to delete equipment: %w", err)
	}

	return nil
}

func (s *EquipmentService) RequestRental(equipmentID, userID uint, startDate, endDate time.Time) (*models.EquipmentRentalResponse, error) {
	var equipment models.Equipment
	if err := s.db.Preload("User").Where("id = ? AND is_available = true", equipmentID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("equipment not found or not available")
		}
		return nil, fmt.Errorf("failed to fetch equipment: %w", err)
	}

	if equipment.UserID == userID {
		return nil, errors.New("cannot rent your own equipment")
	}

	if startDate.Before(time.Now()) {
		return nil, errors.New("start date cannot be in the past")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	var conflictingRentals []models.EquipmentRental
	if err := s.db.Where("equipment_id = ? AND status IN (?) AND ((start_date <= ? AND end_date >= ?) OR (start_date <= ? AND end_date >= ?))",
		equipmentID, []models.RentalStatus{models.RentalStatusApproved, models.RentalStatusActive},
		startDate, startDate, endDate, endDate).Find(&conflictingRentals).Error; err != nil {
		return nil, fmt.Errorf("failed to check conflicting rentals: %w", err)
	}

	if len(conflictingRentals) > 0 {
		return nil, errors.New("equipment is not available for the selected dates")
	}

	days := int(endDate.Sub(startDate).Hours()/24) + 1
	totalPrice := float64(days) * equipment.DailyRentalPrice

	rental := models.EquipmentRental{
		EquipmentID:  equipmentID,
		RenterUserID: userID,
		StartDate:    startDate,
		EndDate:      endDate,
		TotalPrice:   totalPrice,
		Status:       models.RentalStatusRequested,
	}

	if err := s.db.Create(&rental).Error; err != nil {
		return nil, fmt.Errorf("failed to create rental request: %w", err)
	}

	if err := s.db.Preload("Equipment").Preload("Equipment.User").Preload("Renter").First(&rental, rental.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load rental with details: %w", err)
	}

	response := rental.ToResponse()
	return &response, nil
}

func (s *EquipmentService) GetEquipmentRentals(equipmentID, userID uint) ([]models.EquipmentRentalResponse, error) {
	var equipment models.Equipment
	if err := s.db.Where("id = ? AND user_id = ?", equipmentID, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("equipment not found or you don't have permission to view rentals")
		}
		return nil, fmt.Errorf("failed to fetch equipment: %w", err)
	}

	var rentals []models.EquipmentRental
	if err := s.db.Where("equipment_id = ?", equipmentID).
		Preload("Equipment").
		Preload("Equipment.User").
		Preload("Renter").
		Order("created_at DESC").
		Find(&rentals).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch rentals: %w", err)
	}

	responses := make([]models.EquipmentRentalResponse, len(rentals))
	for i, rental := range rentals {
		responses[i] = rental.ToResponse()
	}

	return responses, nil
}

func (s *EquipmentService) UpdateRentalStatus(equipmentID, rentalID, userID uint, status models.RentalStatus) error {
	var equipment models.Equipment
	if err := s.db.Where("id = ? AND user_id = ?", equipmentID, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("equipment not found or you don't have permission to update rentals")
		}
		return fmt.Errorf("failed to fetch equipment: %w", err)
	}

	var rental models.EquipmentRental
	if err := s.db.Where("id = ? AND equipment_id = ?", rentalID, equipmentID).First(&rental).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("rental not found")
		}
		return fmt.Errorf("failed to fetch rental: %w", err)
	}

	if rental.Status != models.RentalStatusRequested && status == models.RentalStatusApproved {
		return errors.New("can only approve requested rentals")
	}

	if err := s.db.Model(&rental).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update rental status: %w", err)
	}

	return nil
}

func (s *EquipmentService) CompleteRental(rentalID, userID uint, returnNotes string) error {
	var rental models.EquipmentRental
	if err := s.db.Preload("Equipment").Where("id = ?", rentalID).First(&rental).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("rental not found")
		}
		return fmt.Errorf("failed to fetch rental: %w", err)
	}

	if rental.Equipment.UserID != userID && rental.RenterUserID != userID {
		return errors.New("you don't have permission to complete this rental")
	}

	if rental.Status != models.RentalStatusActive {
		return errors.New("rental must be active to complete")
	}

	updates := map[string]interface{}{
		"status":       models.RentalStatusCompleted,
		"return_notes": utils.SanitizeString(returnNotes),
	}

	if err := s.db.Model(&rental).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to complete rental: %w", err)
	}

	return nil
}