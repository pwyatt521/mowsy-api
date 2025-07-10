package testutils

import (
	"log"

	"mowsy-api/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Job{},
		&models.JobApplication{},
		&models.Equipment{},
		&models.EquipmentRental{},
		&models.Review{},
		&models.Payment{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans up all tables in the test database
func CleanupTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM payments")
	db.Exec("DELETE FROM reviews")
	db.Exec("DELETE FROM equipment_rentals")
	db.Exec("DELETE FROM equipment")
	db.Exec("DELETE FROM job_applications")
	db.Exec("DELETE FROM jobs")
	db.Exec("DELETE FROM users")
}

// CreateTestUser creates a test user in the database
func CreateTestUser(db *gorm.DB) *models.User {
	user := &models.User{
		Email:                        "test@example.com",
		PasswordHash:                 "$2a$10$test.hash.here",
		FirstName:                    "Test",
		LastName:                     "User",
		Phone:                        "555-123-4567",
		Address:                      "123 Test St",
		City:                         "Test City",
		State:                        "TS",
		ZipCode:                      "12345",
		ElementarySchoolDistrictName: "Test School District",
		ElementarySchoolDistrictCode: "TSD123",
		IsActive:                     true,
		InsuranceVerified:            true,
	}

	if err := db.Create(user).Error; err != nil {
		log.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestJob creates a test job in the database
func CreateTestJob(db *gorm.DB, userID uint) *models.Job {
	job := &models.Job{
		UserID:                       userID,
		Title:                        "Test Lawn Mowing",
		Description:                  "Test description",
		Category:                     models.JobCategoryMowing,
		FixedPrice:                   50.00,
		EstimatedHours:               2.0,
		Address:                      "123 Test St",
		ZipCode:                      "12345",
		ElementarySchoolDistrictName: "Test School District",
		Visibility:                   models.VisibilityZipCode,
		Status:                       models.JobStatusOpen,
	}

	if err := db.Create(job).Error; err != nil {
		log.Fatalf("Failed to create test job: %v", err)
	}

	return job
}

// CreateTestEquipment creates test equipment in the database
func CreateTestEquipment(db *gorm.DB, userID uint) *models.Equipment {
	equipment := &models.Equipment{
		UserID:           userID,
		Name:             "Test Mower",
		Make:             "TestBrand",
		Model:            "TM123",
		Category:         models.EquipmentCategoryMower,
		FuelType:         models.FuelTypeGas,
		PowerType:        models.PowerTypePush,
		DailyRentalPrice: 25.00,
		Description:      "Test mower description",
		ZipCode:          "12345",
		ElementarySchoolDistrictName: "Test School District",
		Visibility:       models.VisibilityZipCode,
		IsAvailable:      true,
	}

	if err := db.Create(equipment).Error; err != nil {
		log.Fatalf("Failed to create test equipment: %v", err)
	}

	return equipment
}