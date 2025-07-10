package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"mowsy-api/internal/routes"
	"mowsy-api/pkg/database"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

var ginLambda *ginadapter.GinLambda
var initialized bool

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
}

func ensureInitialized() error {
	if initialized {
		return nil
	}

	log.Println("Initializing database...")
	// Initialize database
	if err := database.InitDB(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	log.Println("Running auto migrations...")
	// Run auto migrations
	if err := database.AutoMigrate(); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	log.Println("Setting up routes...")
	// Setup routes
	r := routes.SetupRoutes()

	// Create Gin Lambda adapter
	ginLambda = ginadapter.New(r)
	initialized = true
	log.Println("Initialization complete")
	return nil
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Request: %s %s", req.HTTPMethod, req.Path)
	
	// Ensure database and routes are initialized
	if err := ensureInitialized(); err != nil {
		log.Printf("Initialization failed: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error":"Internal server error"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
	
	// Handle the request
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}