package main

import (
	"log"
	"os"
	"waitlist/middleware"
	"waitlist/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading environment variables")
	}

	router := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	router.Use(middleware.CORSMiddleware())

	routes.SetupRoutes(router)
	if err := router.Run(port); err != nil {
		log.Fatal("Unable to start router: ", err)
	}
}
