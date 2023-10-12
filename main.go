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

	routes.SetupRoutes(router)
	if err := router.Run(portAddress); err != nil {
		log.Fatal("Unable to start router: ", err)
	}
}
