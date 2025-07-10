package routes

import (
	"time"

	"mowsy-api/internal/handlers"
	"mowsy-api/internal/middleware"

	"github.com/gin-gonic/gin"
	// swaggerFiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes() *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware(100, time.Minute))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	userHandler := handlers.NewUserHandler()
	jobHandler := handlers.NewJobHandler()
	equipmentHandler := handlers.NewEquipmentHandler()
	paymentHandler := handlers.NewPaymentHandler()
	uploadHandler := handlers.NewUploadHandler()
	adminHandler := handlers.NewAdminHandler()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Swagger documentation endpoint (commented out for Go 1.20 compatibility)
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	api := r.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// Public job listings (with optional auth for user-specific features)
		jobs := api.Group("/jobs")
		jobs.Use(middleware.OptionalAuthMiddleware())
		{
			jobs.GET("", jobHandler.GetJobs)
			jobs.GET("/:id", jobHandler.GetJobByID)
		}

		// Public equipment listings (with optional auth for user-specific features)
		equipment := api.Group("/equipment")
		equipment.Use(middleware.OptionalAuthMiddleware())
		{
			equipment.GET("", equipmentHandler.GetEquipment)
			equipment.GET("/:id", equipmentHandler.GetEquipmentByID)
		}

		// Public user profiles
		users := api.Group("/users")
		{
			users.GET("/:id/profile", userHandler.GetUserPublicProfile)
			users.GET("/:id/reviews", userHandler.GetUserReviews)
		}
	}

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// User management
		users := protected.Group("/users")
		{
			users.GET("/me", userHandler.GetCurrentUser)
			users.PUT("/me", userHandler.UpdateCurrentUser)
			users.POST("/me/insurance", userHandler.UploadInsuranceDocument)
		}

		// Job management
		jobs := protected.Group("/jobs")
		{
			jobs.POST("", jobHandler.CreateJob)
			jobs.PUT("/:id", jobHandler.UpdateJob)
			jobs.DELETE("/:id", jobHandler.DeleteJob)
			jobs.POST("/:id/apply", jobHandler.ApplyForJob)
			jobs.GET("/:id/applications", jobHandler.GetJobApplications)
			jobs.PUT("/:id/applications/:app_id", jobHandler.UpdateApplicationStatus)
			jobs.POST("/:id/complete", middleware.InsuranceRequiredMiddleware(), jobHandler.CompleteJob)
		}

		// Equipment management
		equipment := protected.Group("/equipment")
		{
			equipment.POST("", equipmentHandler.CreateEquipment)
			equipment.PUT("/:id", equipmentHandler.UpdateEquipment)
			equipment.DELETE("/:id", equipmentHandler.DeleteEquipment)
			equipment.POST("/:id/rent", equipmentHandler.RequestRental)
			equipment.GET("/:id/rentals", equipmentHandler.GetEquipmentRentals)
			equipment.PUT("/:id/rentals/:rental_id", equipmentHandler.UpdateRentalStatus)
			equipment.POST("/rentals/:rental_id/complete", middleware.InsuranceRequiredMiddleware(), equipmentHandler.CompleteRental)
		}

		// Payment processing
		payments := protected.Group("/payments")
		{
			payments.POST("/create-intent", paymentHandler.CreatePaymentIntent)
			payments.POST("/confirm", paymentHandler.ConfirmPayment)
			payments.GET("/history", paymentHandler.GetPaymentHistory)
			payments.GET("/:id", paymentHandler.GetPaymentByID)
		}

		// File upload
		upload := protected.Group("/upload")
		{
			upload.POST("/image", uploadHandler.UploadImage)
			upload.POST("/presigned-url", uploadHandler.GetPresignedUploadURL)
			upload.DELETE("/file", uploadHandler.DeleteFile)
		}
	}

	// Admin routes (require admin API key)
	admin := api.Group("/admin")
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/stats", adminHandler.GetStats)
		admin.GET("/users", adminHandler.GetUsers)
		admin.PUT("/users/:id/deactivate", adminHandler.DeactivateUser)
		admin.PUT("/users/:id/activate", adminHandler.ActivateUser)
		admin.PUT("/users/:id/verify-insurance", adminHandler.VerifyInsurance)
		admin.DELETE("/jobs/:id", adminHandler.RemoveJob)
		admin.DELETE("/equipment/:id", adminHandler.RemoveEquipment)
	}

	return r
}