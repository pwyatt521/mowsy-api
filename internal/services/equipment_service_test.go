package services

import (
	"testing"

	"mowsy-api/internal/models"
	"mowsy-api/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupEquipmentService() (*EquipmentService, *gorm.DB) {
	db := testutils.SetupTestDB()
	service := &EquipmentService{db: db}
	return service, db
}

func TestEquipmentService_GetEquipmentByUserID(t *testing.T) {
	service, db := setupEquipmentService()
	defer testutils.CleanupTestDB(db)

	// Create test users
	user1 := testutils.CreateTestUser(db)
	user2 := &models.User{
		Email:     "user2@example.com",
		FirstName: "User2",
		LastName:  "Test",
		IsActive:  true,
	}
	err := db.Create(user2).Error
	require.NoError(t, err)

	// Create equipment for user1
	equipment1 := &models.Equipment{
		UserID:           user1.ID,
		Name:             "Lawn Mower 1",
		Category:         models.EquipmentCategoryMower,
		FuelType:         models.FuelTypeGas,
		PowerType:        models.PowerTypePush,
		DailyRentalPrice: 30.00,
		IsAvailable:      true,
		Visibility:       models.VisibilityZipCode,
	}
	err = db.Create(equipment1).Error
	require.NoError(t, err)

	equipment2 := &models.Equipment{
		UserID:           user1.ID,
		Name:             "Weed Whacker 1",
		Category:         models.EquipmentCategoryWeedWhacker,
		FuelType:         models.FuelTypeElectric,
		PowerType:        models.PowerTypeCorded,
		DailyRentalPrice: 15.00,
		IsAvailable:      true,
		Visibility:       models.VisibilityZipCode,
	}
	err = db.Create(equipment2).Error
	require.NoError(t, err)

	// Create unavailable equipment for user1
	equipment3 := &models.Equipment{
		UserID:           user1.ID,
		Name:             "Edger 1",
		Category:         models.EquipmentCategoryEdger,
		FuelType:         models.FuelTypeGas,
		PowerType:        models.PowerTypeGas,
		DailyRentalPrice: 20.00,
		IsAvailable:      false,
		Visibility:       models.VisibilityZipCode,
	}
	err = db.Create(equipment3).Error
	require.NoError(t, err)
	
	// Explicitly update IsAvailable to false since GORM has default:true
	err = db.Model(equipment3).Update("is_available", false).Error
	require.NoError(t, err)

	// Create equipment for user2
	equipment4 := &models.Equipment{
		UserID:           user2.ID,
		Name:             "User2 Equipment",
		Category:         models.EquipmentCategoryMower,
		DailyRentalPrice: 40.00,
		IsAvailable:      true,
		Visibility:       models.VisibilityZipCode,
	}
	err = db.Create(equipment4).Error
	require.NoError(t, err)

	t.Run("GetAllUserEquipment", func(t *testing.T) {
		filters := EquipmentFilters{}

		equipment, err := service.GetEquipmentByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 3)
		assert.Equal(t, equipment3.Name, equipment[0].Name) // Most recent first (by created_at DESC)
		assert.Equal(t, equipment2.Name, equipment[1].Name)
		assert.Equal(t, equipment1.Name, equipment[2].Name)
	})

	t.Run("FilterByCategory", func(t *testing.T) {
		filters := EquipmentFilters{
			Category: models.EquipmentCategoryWeedWhacker,
		}

		equipment, err := service.GetEquipmentByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 1)
		assert.Equal(t, equipment2.Name, equipment[0].Name)
	})

	t.Run("FilterByAvailability", func(t *testing.T) {
		available := true
		filters := EquipmentFilters{
			IsAvailable: &available,
		}

		equipment, err := service.GetEquipmentByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 2)
		// Verify only available equipment is returned
		for _, eq := range equipment {
			assert.True(t, eq.IsAvailable, "Equipment %s should be available", eq.Name)
		}
	})

	t.Run("FilterByUnavailability", func(t *testing.T) {
		available := false
		filters := EquipmentFilters{
			IsAvailable: &available,
		}

		equipment, err := service.GetEquipmentByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 1)
		assert.Equal(t, equipment3.Name, equipment[0].Name)
		assert.False(t, equipment[0].IsAvailable)
	})

	t.Run("FilterByFuelType", func(t *testing.T) {
		filters := EquipmentFilters{
			FuelType: models.FuelTypeGas,
		}

		equipment, err := service.GetEquipmentByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 2)
		assert.Equal(t, equipment3.Name, equipment[0].Name) // Most recent first
		assert.Equal(t, equipment1.Name, equipment[1].Name)
	})

	t.Run("FilterByPowerType", func(t *testing.T) {
		filters := EquipmentFilters{
			PowerType: models.PowerTypeCorded,
		}

		equipment, err := service.GetEquipmentByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 1)
		assert.Equal(t, equipment2.Name, equipment[0].Name)
	})

	t.Run("Pagination", func(t *testing.T) {
		filters := EquipmentFilters{
			Page:  1,
			Limit: 1,
		}

		equipment, err := service.GetEquipmentByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 1)
		assert.Equal(t, equipment3.Name, equipment[0].Name) // Most recent first
	})

	t.Run("EmptyResultForOtherUser", func(t *testing.T) {
		filters := EquipmentFilters{}

		equipment, err := service.GetEquipmentByUserID(99999, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 0)
	})

	t.Run("DefaultPagination", func(t *testing.T) {
		filters := EquipmentFilters{
			Page:  0, // Should default to 1
			Limit: 0, // Should default to 20
		}

		equipment, err := service.GetEquipmentByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 3)
	})

	t.Run("CombinedFilters", func(t *testing.T) {
		available := true
		filters := EquipmentFilters{
			Category:    models.EquipmentCategoryMower,
			IsAvailable: &available,
		}

		equipment, err := service.GetEquipmentByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, equipment, 1)
		assert.Equal(t, equipment1.Name, equipment[0].Name)
		assert.Equal(t, models.EquipmentCategoryMower, equipment[0].Category)
		assert.True(t, equipment[0].IsAvailable)
	})
}

func TestEquipmentService_GetEquipmentWithFilter(t *testing.T) {
	service, db := setupEquipmentService()
	defer testutils.CleanupTestDB(db)

	// Create test users in different locations
	user1 := &models.User{
		Email:                        "user1@example.com",
		FirstName:                    "User1",
		LastName:                     "Test",
		ZipCode:                      "12345",
		ElementarySchoolDistrictName: "District A",
		IsActive:                     true,
	}
	err := db.Create(user1).Error
	require.NoError(t, err)

	user2 := &models.User{
		Email:                        "user2@example.com",
		FirstName:                    "User2",
		LastName:                     "Test",
		ZipCode:                      "12345", // Same zip as user1
		ElementarySchoolDistrictName: "District B",
		IsActive:                     true,
	}
	err = db.Create(user2).Error
	require.NoError(t, err)

	user3 := &models.User{
		Email:                        "user3@example.com",
		FirstName:                    "User3",
		LastName:                     "Test",
		ZipCode:                      "67890", // Different zip
		ElementarySchoolDistrictName: "District A", // Same district as user1
		IsActive:                     true,
	}
	err = db.Create(user3).Error
	require.NoError(t, err)

	user4 := &models.User{
		Email:                        "user4@example.com",
		FirstName:                    "User4",
		LastName:                     "Test",
		ZipCode:                      "99999", // Different zip
		ElementarySchoolDistrictName: "District C", // Different district
		IsActive:                     true,
	}
	err = db.Create(user4).Error
	require.NoError(t, err)

	// Create equipment with different visibility settings
	// User1's own equipment (should be excluded from filtered results)
	equipment1 := &models.Equipment{
		UserID:                       user1.ID,
		Name:                         "User1 Equipment - Should be excluded",
		Category:                     models.EquipmentCategoryMower,
		DailyRentalPrice:             50.00,
		IsAvailable:                  true,
		ZipCode:                      user1.ZipCode,
		ElementarySchoolDistrictName: user1.ElementarySchoolDistrictName,
		Visibility:                   models.VisibilityZipCode,
	}
	err = db.Create(equipment1).Error
	require.NoError(t, err)

	// User2's equipment - same zip as user1, should be visible with zip_code visibility
	equipment2 := &models.Equipment{
		UserID:                       user2.ID,
		Name:                         "User2 Equipment - Same Zip",
		Category:                     models.EquipmentCategoryWeedWhacker,
		DailyRentalPrice:             30.00,
		IsAvailable:                  true,
		ZipCode:                      user2.ZipCode,
		ElementarySchoolDistrictName: user2.ElementarySchoolDistrictName,
		Visibility:                   models.VisibilityZipCode,
	}
	err = db.Create(equipment2).Error
	require.NoError(t, err)

	// User3's equipment - same district as user1, should be visible with school_district visibility
	equipment3 := &models.Equipment{
		UserID:                       user3.ID,
		Name:                         "User3 Equipment - Same District",
		Category:                     models.EquipmentCategoryEdger,
		DailyRentalPrice:             40.00,
		IsAvailable:                  true,
		ZipCode:                      user3.ZipCode,
		ElementarySchoolDistrictName: user3.ElementarySchoolDistrictName,
		Visibility:                   models.VisibilitySchoolDistrict,
	}
	err = db.Create(equipment3).Error
	require.NoError(t, err)

	// User4's equipment - different zip and district, should NOT be visible
	equipment4 := &models.Equipment{
		UserID:                       user4.ID,
		Name:                         "User4 Equipment - Should NOT be visible",
		Category:                     models.EquipmentCategoryMower,
		DailyRentalPrice:             60.00,
		IsAvailable:                  true,
		ZipCode:                      user4.ZipCode,
		ElementarySchoolDistrictName: user4.ElementarySchoolDistrictName,
		Visibility:                   models.VisibilityZipCode,
	}
	err = db.Create(equipment4).Error
	require.NoError(t, err)

	t.Run("FilterEnabled_ExcludesOwnEquipment_AppliesVisibilityFiltering", func(t *testing.T) {
		filterEnabled := true
		filters := EquipmentFilters{
			Filter: &filterEnabled,
		}

		equipment, err := service.GetEquipmentWithUser(filters, &user1.ID)

		require.NoError(t, err)
		assert.Len(t, equipment, 2) // Should see equipment2 (same zip) and equipment3 (same district), but not equipment1 (own) or equipment4 (different location)
		
		// Check that we get the correct equipment
		equipmentNames := make([]string, len(equipment))
		for i, eq := range equipment {
			equipmentNames[i] = eq.Name
		}
		assert.Contains(t, equipmentNames, "User2 Equipment - Same Zip")
		assert.Contains(t, equipmentNames, "User3 Equipment - Same District")
		assert.NotContains(t, equipmentNames, "User1 Equipment - Should be excluded")
		assert.NotContains(t, equipmentNames, "User4 Equipment - Should NOT be visible")
	})

	t.Run("FilterDisabled_ShowsAllEquipment", func(t *testing.T) {
		filterEnabled := false
		filters := EquipmentFilters{
			Filter: &filterEnabled,
		}

		equipment, err := service.GetEquipmentWithUser(filters, &user1.ID)

		require.NoError(t, err)
		assert.Len(t, equipment, 4) // Should see all equipment when filter is disabled
	})

	t.Run("FilterNotSet_ShowsAllEquipment", func(t *testing.T) {
		filters := EquipmentFilters{
			// Filter not set
		}

		equipment, err := service.GetEquipmentWithUser(filters, &user1.ID)

		require.NoError(t, err)
		assert.Len(t, equipment, 4) // Should see all equipment when filter is not set
	})

	t.Run("FilterEnabled_NoUserID_ShowsAllEquipment", func(t *testing.T) {
		filterEnabled := true
		filters := EquipmentFilters{
			Filter: &filterEnabled,
		}

		equipment, err := service.GetEquipmentWithUser(filters, nil)

		require.NoError(t, err)
		assert.Len(t, equipment, 4) // Should see all equipment when no userID provided
	})

	t.Run("FilterEnabled_UnavailableEquipmentExcluded", func(t *testing.T) {
		// Create unavailable equipment that should be excluded
		unavailableEquipment := &models.Equipment{
			UserID:                       user2.ID,
			Name:                         "User2 Unavailable Equipment",
			Category:                     models.EquipmentCategoryMower,
			DailyRentalPrice:             25.00,
			IsAvailable:                  false,
			ZipCode:                      user2.ZipCode,
			ElementarySchoolDistrictName: user2.ElementarySchoolDistrictName,
			Visibility:                   models.VisibilityZipCode,
		}
		err = db.Create(unavailableEquipment).Error
		require.NoError(t, err)
		
		// Explicitly update IsAvailable to false since GORM has default:true
		err = db.Model(unavailableEquipment).Update("is_available", false).Error
		require.NoError(t, err)

		filterEnabled := true
		filters := EquipmentFilters{
			Filter: &filterEnabled,
		}

		equipment, err := service.GetEquipmentWithUser(filters, &user1.ID)

		require.NoError(t, err)
		// Should still see only the 2 available equipment items, unavailable one should be excluded by default IsAvailable filter
		assert.Len(t, equipment, 2)
		
		equipmentNames := make([]string, len(equipment))
		for i, eq := range equipment {
			equipmentNames[i] = eq.Name
		}
		assert.NotContains(t, equipmentNames, "User2 Unavailable Equipment")
	})
}