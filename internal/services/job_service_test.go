package services

import (
	"testing"

	"mowsy-api/internal/models"
	"mowsy-api/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupJobService() (*JobService, *gorm.DB) {
	db := testutils.SetupTestDB()
	service := &JobService{db: db}
	return service, db
}

func TestJobService_CreateJob(t *testing.T) {
	service, db := setupJobService()
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := testutils.CreateTestUser(db)

	t.Run("ValidJobCreation", func(t *testing.T) {
		req := CreateJobRequest{
			Title:          "Test Lawn Mowing",
			Description:    "Need lawn mowed",
			Category:       models.JobCategoryMowing,
			FixedPrice:     50.00,
			EstimatedHours: 2.0,
			Address:        "123 Test St",
			Visibility:     models.VisibilityZipCode,
		}

		job, err := service.CreateJob(user.ID, req)

		require.NoError(t, err)
		assert.NotNil(t, job)
		assert.Equal(t, req.Title, job.Title)
		assert.Equal(t, req.Description, job.Description)
		assert.Equal(t, req.Category, job.Category)
		assert.Equal(t, req.FixedPrice, job.FixedPrice)
		assert.Equal(t, models.JobStatusOpen, job.Status)
		assert.Equal(t, user.ID, job.User.ID)
	})

	t.Run("InvalidPrice", func(t *testing.T) {
		req := CreateJobRequest{
			Title:       "Test Job",
			Category:    models.JobCategoryMowing,
			FixedPrice:  0, // Invalid price
			Visibility:  models.VisibilityZipCode,
		}

		job, err := service.CreateJob(user.ID, req)

		assert.Error(t, err)
		assert.Nil(t, job)
		assert.Contains(t, err.Error(), "fixed price must be greater than 0")
	})

	t.Run("NegativeEstimatedHours", func(t *testing.T) {
		req := CreateJobRequest{
			Title:          "Test Job",
			Category:       models.JobCategoryMowing,
			FixedPrice:     50.00,
			EstimatedHours: -1, // Invalid hours
			Visibility:     models.VisibilityZipCode,
		}

		job, err := service.CreateJob(user.ID, req)

		assert.Error(t, err)
		assert.Nil(t, job)
		assert.Contains(t, err.Error(), "estimated hours cannot be negative")
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		req := CreateJobRequest{
			Title:      "Test Job",
			Category:   models.JobCategoryMowing,
			FixedPrice: 50.00,
			Visibility: models.VisibilityZipCode,
		}

		job, err := service.CreateJob(99999, req)

		assert.Error(t, err)
		assert.Nil(t, job)
		assert.Contains(t, err.Error(), "failed to get user")
	})
}

func TestJobService_GetJobs(t *testing.T) {
	service, db := setupJobService()
	defer testutils.CleanupTestDB(db)

	// Create test users and jobs
	user1 := testutils.CreateTestUser(db)
	user2 := &models.User{
		Email:     "user2@example.com",
		FirstName: "User2",
		LastName:  "Test",
		ZipCode:   "54321",
		ElementarySchoolDistrictName: "Another School District",
		IsActive:  true,
	}
	err := db.Create(user2).Error
	require.NoError(t, err)

	job1 := testutils.CreateTestJob(db, user1.ID)
	job2 := &models.Job{
		UserID:      user2.ID,
		Title:       "Weeding Job",
		Category:    models.JobCategoryWeeding,
		FixedPrice:  30.00,
		ZipCode:     "54321",
		ElementarySchoolDistrictName: "Another School District",
		Visibility:  models.VisibilityZipCode,
		Status:      models.JobStatusOpen,
	}
	err = db.Create(job2).Error
	require.NoError(t, err)

	t.Run("GetAllJobs", func(t *testing.T) {
		filters := JobFilters{}

		jobs, err := service.GetJobs(filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 2)
	})

	t.Run("FilterByZipCode", func(t *testing.T) {
		filters := JobFilters{
			ZipCode: "12345",
		}

		jobs, err := service.GetJobs(filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, job1.Title, jobs[0].Title)
	})

	t.Run("FilterByCategory", func(t *testing.T) {
		filters := JobFilters{
			Category: models.JobCategoryWeeding,
		}

		jobs, err := service.GetJobs(filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, job2.Title, jobs[0].Title)
	})

	t.Run("FilterBySchoolDistrict", func(t *testing.T) {
		filters := JobFilters{
			District: "Test School District",
		}

		jobs, err := service.GetJobs(filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, job1.Title, jobs[0].Title)
	})

	t.Run("FilterByPriceRange", func(t *testing.T) {
		minPrice := 40.0
		maxPrice := 60.0
		filters := JobFilters{
			MinPrice: &minPrice,
			MaxPrice: &maxPrice,
		}

		jobs, err := service.GetJobs(filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, job1.Title, jobs[0].Title)
	})

	t.Run("Pagination", func(t *testing.T) {
		filters := JobFilters{
			Page:  1,
			Limit: 1,
		}

		jobs, err := service.GetJobs(filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 1)
	})
}

func TestJobService_UpdateJob(t *testing.T) {
	service, db := setupJobService()
	defer testutils.CleanupTestDB(db)

	// Create test user and job
	user := testutils.CreateTestUser(db)
	job := testutils.CreateTestJob(db, user.ID)

	t.Run("ValidUpdate", func(t *testing.T) {
		newPrice := 75.0
		req := UpdateJobRequest{
			Title:      "Updated Job Title",
			FixedPrice: &newPrice,
		}

		updatedJob, err := service.UpdateJob(job.ID, user.ID, req)

		require.NoError(t, err)
		assert.NotNil(t, updatedJob)
		assert.Equal(t, "Updated Job Title", updatedJob.Title)
		assert.Equal(t, 75.0, updatedJob.FixedPrice)
	})

	t.Run("InvalidPrice", func(t *testing.T) {
		invalidPrice := -10.0
		req := UpdateJobRequest{
			FixedPrice: &invalidPrice,
		}

		updatedJob, err := service.UpdateJob(job.ID, user.ID, req)

		assert.Error(t, err)
		assert.Nil(t, updatedJob)
		assert.Contains(t, err.Error(), "fixed price must be greater than 0")
	})

	t.Run("UnauthorizedUpdate", func(t *testing.T) {
		req := UpdateJobRequest{
			Title: "Unauthorized Update",
		}

		updatedJob, err := service.UpdateJob(job.ID, 99999, req)

		assert.Error(t, err)
		assert.Nil(t, updatedJob)
		assert.Contains(t, err.Error(), "job not found or you don't have permission")
	})

	t.Run("UpdateNonOpenJob", func(t *testing.T) {
		// Create a completed job
		completedJob := &models.Job{
			UserID:     user.ID,
			Title:      "Completed Job",
			Category:   models.JobCategoryMowing,
			FixedPrice: 50.00,
			Status:     models.JobStatusCompleted,
			Visibility: models.VisibilityZipCode,
		}
		err := db.Create(completedJob).Error
		require.NoError(t, err)

		req := UpdateJobRequest{
			Title: "Cannot Update",
		}

		updatedJob, err := service.UpdateJob(completedJob.ID, user.ID, req)

		assert.Error(t, err)
		assert.Nil(t, updatedJob)
		assert.Contains(t, err.Error(), "cannot update job that is not open")
	})
}

func TestJobService_ApplyForJob(t *testing.T) {
	service, db := setupJobService()
	defer testutils.CleanupTestDB(db)

	// Create test users and job
	jobOwner := testutils.CreateTestUser(db)
	applicant := &models.User{
		Email:     "applicant@example.com",
		FirstName: "Applicant",
		LastName:  "User",
		IsActive:  true,
	}
	err := db.Create(applicant).Error
	require.NoError(t, err)

	job := testutils.CreateTestJob(db, jobOwner.ID)

	t.Run("ValidApplication", func(t *testing.T) {
		message := "I would like to apply for this job"

		application, err := service.ApplyForJob(job.ID, applicant.ID, message)

		require.NoError(t, err)
		assert.NotNil(t, application)
		assert.Equal(t, job.ID, application.JobID)
		assert.Equal(t, message, application.Message)
		assert.Equal(t, models.ApplicationStatusPending, application.Status)
	})

	t.Run("ApplyForOwnJob", func(t *testing.T) {
		message := "Applying for my own job"

		application, err := service.ApplyForJob(job.ID, jobOwner.ID, message)

		assert.Error(t, err)
		assert.Nil(t, application)
		assert.Contains(t, err.Error(), "cannot apply for your own job")
	})

	t.Run("DuplicateApplication", func(t *testing.T) {
		message := "Duplicate application"

		// Apply again with the same user
		application, err := service.ApplyForJob(job.ID, applicant.ID, message)

		assert.Error(t, err)
		assert.Nil(t, application)
		assert.Contains(t, err.Error(), "you have already applied for this job")
	})

	t.Run("ApplyForNonexistentJob", func(t *testing.T) {
		message := "Applying for nonexistent job"

		application, err := service.ApplyForJob(99999, applicant.ID, message)

		assert.Error(t, err)
		assert.Nil(t, application)
		assert.Contains(t, err.Error(), "job not found or not accepting applications")
	})
}

func TestJobService_UpdateApplicationStatus(t *testing.T) {
	service, db := setupJobService()
	defer testutils.CleanupTestDB(db)

	// Create test users, job, and application
	jobOwner := testutils.CreateTestUser(db)
	applicant := &models.User{
		Email:     "applicant@example.com",
		FirstName: "Applicant",
		LastName:  "User",
		IsActive:  true,
	}
	err := db.Create(applicant).Error
	require.NoError(t, err)

	job := testutils.CreateTestJob(db, jobOwner.ID)
	
	application := &models.JobApplication{
		JobID:   job.ID,
		UserID:  applicant.ID,
		Message: "Test application",
		Status:  models.ApplicationStatusPending,
	}
	err = db.Create(application).Error
	require.NoError(t, err)

	t.Run("AcceptApplication", func(t *testing.T) {
		err := service.UpdateApplicationStatus(job.ID, application.ID, jobOwner.ID, models.ApplicationStatusAccepted)

		require.NoError(t, err)

		// Verify application status is updated
		var updatedApp models.JobApplication
		err = db.First(&updatedApp, application.ID).Error
		require.NoError(t, err)
		assert.Equal(t, models.ApplicationStatusAccepted, updatedApp.Status)

		// Verify job status is updated to in_progress
		var updatedJob models.Job
		err = db.First(&updatedJob, job.ID).Error
		require.NoError(t, err)
		assert.Equal(t, models.JobStatusInProgress, updatedJob.Status)
	})

	t.Run("UnauthorizedStatusUpdate", func(t *testing.T) {
		err := service.UpdateApplicationStatus(job.ID, application.ID, 99999, models.ApplicationStatusRejected)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job not found or you don't have permission")
	})
}

func TestJobService_CompleteJob(t *testing.T) {
	service, db := setupJobService()
	defer testutils.CleanupTestDB(db)

	// Create test user and job
	user := testutils.CreateTestUser(db)
	job := testutils.CreateTestJob(db, user.ID)

	// Set job to in_progress status
	err := db.Model(job).Update("status", models.JobStatusInProgress).Error
	require.NoError(t, err)

	t.Run("ValidCompletion", func(t *testing.T) {
		imageUrls := []string{
			"https://s3.amazonaws.com/bucket/before.jpg",
			"https://s3.amazonaws.com/bucket/after.jpg",
		}

		err := service.CompleteJob(job.ID, user.ID, imageUrls)

		require.NoError(t, err)

		// Verify job status is updated
		var updatedJob models.Job
		err = db.First(&updatedJob, job.ID).Error
		require.NoError(t, err)
		assert.Equal(t, models.JobStatusCompleted, updatedJob.Status)
		assert.Equal(t, models.StringArray(imageUrls), updatedJob.CompletionImageUrls)
	})

	t.Run("CompleteJobNotInProgress", func(t *testing.T) {
		// Create an open job
		openJob := &models.Job{
			UserID:     user.ID,
			Title:      "Open Job",
			Category:   models.JobCategoryMowing,
			FixedPrice: 50.00,
			Status:     models.JobStatusOpen,
			Visibility: models.VisibilityZipCode,
		}
		err := db.Create(openJob).Error
		require.NoError(t, err)

		imageUrls := []string{"https://s3.amazonaws.com/bucket/image.jpg"}

		err = service.CompleteJob(openJob.ID, user.ID, imageUrls)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job must be in progress to complete")
	})

	t.Run("UnauthorizedCompletion", func(t *testing.T) {
		imageUrls := []string{"https://s3.amazonaws.com/bucket/image.jpg"}

		err := service.CompleteJob(job.ID, 99999, imageUrls)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job not found or you don't have permission")
	})
}

func TestJobService_DeleteJob(t *testing.T) {
	service, db := setupJobService()
	defer testutils.CleanupTestDB(db)

	// Create test user and job
	user := testutils.CreateTestUser(db)
	job := testutils.CreateTestJob(db, user.ID)

	t.Run("ValidDeletion", func(t *testing.T) {
		err := service.DeleteJob(job.ID, user.ID)

		require.NoError(t, err)

		// Verify job is deleted
		var deletedJob models.Job
		err = db.First(&deletedJob, job.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("UnauthorizedDeletion", func(t *testing.T) {
		anotherJob := testutils.CreateTestJob(db, user.ID)

		err := service.DeleteJob(anotherJob.ID, 99999)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job not found or you don't have permission")
	})

	t.Run("DeleteNonOpenJob", func(t *testing.T) {
		// Create a completed job
		completedJob := &models.Job{
			UserID:     user.ID,
			Title:      "Completed Job",
			Category:   models.JobCategoryMowing,
			FixedPrice: 50.00,
			Status:     models.JobStatusCompleted,
			Visibility: models.VisibilityZipCode,
		}
		err := db.Create(completedJob).Error
		require.NoError(t, err)

		err = service.DeleteJob(completedJob.ID, user.ID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete job that is not open")
	})
}

func TestJobService_GetJobsByUserID(t *testing.T) {
	service, db := setupJobService()
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

	// Create jobs for user1
	job1 := &models.Job{
		UserID:      user1.ID,
		Title:       "Mowing Job 1",
		Category:    models.JobCategoryMowing,
		FixedPrice:  50.00,
		Status:      models.JobStatusOpen,
		Visibility:  models.VisibilityZipCode,
	}
	err = db.Create(job1).Error
	require.NoError(t, err)

	job2 := &models.Job{
		UserID:      user1.ID,
		Title:       "Weeding Job 1",
		Category:    models.JobCategoryWeeding,
		FixedPrice:  30.00,
		Status:      models.JobStatusCompleted,
		Visibility:  models.VisibilityZipCode,
	}
	err = db.Create(job2).Error
	require.NoError(t, err)

	// Create job for user2
	job3 := &models.Job{
		UserID:      user2.ID,
		Title:       "User2 Job",
		Category:    models.JobCategoryMowing,
		FixedPrice:  40.00,
		Status:      models.JobStatusOpen,
		Visibility:  models.VisibilityZipCode,
	}
	err = db.Create(job3).Error
	require.NoError(t, err)

	t.Run("GetAllUserJobs", func(t *testing.T) {
		filters := JobFilters{}

		jobs, err := service.GetJobsByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 2)
		assert.Equal(t, job2.Title, jobs[0].Title) // Most recent first (by created_at DESC)
		assert.Equal(t, job1.Title, jobs[1].Title)
	})

	t.Run("FilterByStatus", func(t *testing.T) {
		filters := JobFilters{
			Status: models.JobStatusOpen,
		}

		jobs, err := service.GetJobsByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, job1.Title, jobs[0].Title)
	})

	t.Run("FilterByCategory", func(t *testing.T) {
		filters := JobFilters{
			Category: models.JobCategoryWeeding,
		}

		jobs, err := service.GetJobsByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, job2.Title, jobs[0].Title)
	})

	t.Run("Pagination", func(t *testing.T) {
		filters := JobFilters{
			Page:  1,
			Limit: 1,
		}

		jobs, err := service.GetJobsByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, job2.Title, jobs[0].Title) // Most recent first
	})

	t.Run("EmptyResultForOtherUser", func(t *testing.T) {
		filters := JobFilters{}

		jobs, err := service.GetJobsByUserID(99999, filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 0)
	})

	t.Run("DefaultPagination", func(t *testing.T) {
		filters := JobFilters{
			Page:  0, // Should default to 1
			Limit: 0, // Should default to 20
		}

		jobs, err := service.GetJobsByUserID(user1.ID, filters)

		require.NoError(t, err)
		assert.Len(t, jobs, 2)
	})
}

func TestJobService_GetJobsWithFilter(t *testing.T) {
	service, db := setupJobService()
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

	// Create jobs with different visibility settings
	// User1's own job (should be excluded from filtered results)
	job1 := &models.Job{
		UserID:                       user1.ID,
		Title:                        "User1 Job - Should be excluded",
		Category:                     models.JobCategoryMowing,
		FixedPrice:                   50.00,
		Status:                       models.JobStatusOpen,
		ZipCode:                      user1.ZipCode,
		ElementarySchoolDistrictName: user1.ElementarySchoolDistrictName,
		Visibility:                   models.VisibilityZipCode,
	}
	err = db.Create(job1).Error
	require.NoError(t, err)

	// User2's job - same zip as user1, should be visible with zip_code visibility
	job2 := &models.Job{
		UserID:                       user2.ID,
		Title:                        "User2 Job - Same Zip",
		Category:                     models.JobCategoryWeeding,
		FixedPrice:                   30.00,
		Status:                       models.JobStatusOpen,
		ZipCode:                      user2.ZipCode,
		ElementarySchoolDistrictName: user2.ElementarySchoolDistrictName,
		Visibility:                   models.VisibilityZipCode,
	}
	err = db.Create(job2).Error
	require.NoError(t, err)

	// User3's job - same district as user1, should be visible with school_district visibility
	job3 := &models.Job{
		UserID:                       user3.ID,
		Title:                        "User3 Job - Same District",
		Category:                     models.JobCategoryTrimming,
		FixedPrice:                   40.00,
		Status:                       models.JobStatusOpen,
		ZipCode:                      user3.ZipCode,
		ElementarySchoolDistrictName: user3.ElementarySchoolDistrictName,
		Visibility:                   models.VisibilitySchoolDistrict,
	}
	err = db.Create(job3).Error
	require.NoError(t, err)

	// User4's job - different zip and district, should NOT be visible
	job4 := &models.Job{
		UserID:                       user4.ID,
		Title:                        "User4 Job - Should NOT be visible",
		Category:                     models.JobCategoryMowing,
		FixedPrice:                   60.00,
		Status:                       models.JobStatusOpen,
		ZipCode:                      user4.ZipCode,
		ElementarySchoolDistrictName: user4.ElementarySchoolDistrictName,
		Visibility:                   models.VisibilityZipCode,
	}
	err = db.Create(job4).Error
	require.NoError(t, err)

	t.Run("FilterEnabled_ExcludesOwnJobs_AppliesVisibilityFiltering", func(t *testing.T) {
		filterEnabled := true
		filters := JobFilters{
			Filter: &filterEnabled,
		}

		jobs, err := service.GetJobsWithUser(filters, &user1.ID)

		require.NoError(t, err)
		assert.Len(t, jobs, 2) // Should see job2 (same zip) and job3 (same district), but not job1 (own) or job4 (different location)
		
		// Check that we get the correct jobs
		jobTitles := make([]string, len(jobs))
		for i, job := range jobs {
			jobTitles[i] = job.Title
		}
		assert.Contains(t, jobTitles, "User2 Job - Same Zip")
		assert.Contains(t, jobTitles, "User3 Job - Same District")
		assert.NotContains(t, jobTitles, "User1 Job - Should be excluded")
		assert.NotContains(t, jobTitles, "User4 Job - Should NOT be visible")
	})

	t.Run("FilterDisabled_ShowsAllJobs", func(t *testing.T) {
		filterEnabled := false
		filters := JobFilters{
			Filter: &filterEnabled,
		}

		jobs, err := service.GetJobsWithUser(filters, &user1.ID)

		require.NoError(t, err)
		assert.Len(t, jobs, 4) // Should see all jobs when filter is disabled
	})

	t.Run("FilterNotSet_ShowsAllJobs", func(t *testing.T) {
		filters := JobFilters{
			// Filter not set
		}

		jobs, err := service.GetJobsWithUser(filters, &user1.ID)

		require.NoError(t, err)
		assert.Len(t, jobs, 4) // Should see all jobs when filter is not set
	})

	t.Run("FilterEnabled_NoUserID_ShowsAllJobs", func(t *testing.T) {
		filterEnabled := true
		filters := JobFilters{
			Filter: &filterEnabled,
		}

		jobs, err := service.GetJobsWithUser(filters, nil)

		require.NoError(t, err)
		assert.Len(t, jobs, 4) // Should see all jobs when no userID provided
	})
}