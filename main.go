package main

import (
	"fmt"
	"log"
	"os"
	"waitlist/middleware"
	"waitlist/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	gin.SetMode(gin.ReleaseMode)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading environment variables")
	}

	router := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	portAddress := fmt.Sprintf(":%s", port)

	router.Use(middleware.CORSMiddleware())

	// Initialize AuthConn
	privateKey := os.Getenv("PRIVATE_KEY")
	publicKey := os.Getenv("PUBLIC_KEY")
	if privateKey == "" || publicKey == "" {
		log.Fatal("environment variable not set")
	}
	authConn := middleware.NewAuthConn(privateKey, publicKey)

	// Setup routes with AuthMiddleware
	routes.SetupRoutes(router, authConn)

	if err := router.Run(portAddress); err != nil {
		log.Fatal("Unable to start router: ", err)
	}
}
