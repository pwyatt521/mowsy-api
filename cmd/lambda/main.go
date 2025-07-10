package main

import (
	"context"
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

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run auto migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run auto migrations: %v", err)
	}

	// Setup routes
	r := routes.SetupRoutes()

	// Create Gin Lambda adapter
	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Request: %s %s", req.HTTPMethod, req.Path)
	
	// Handle the request
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}