package routes

import (
	"waitlist/controllers"
	"waitlist/db"
	"waitlist/lib/emailclient/postmark"
	"waitlist/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, authConn *middleware.AuthConn) {
	email := postmark.New()
	wt := controllers.NewWaitlist(db.ConnectDatabase(), email, authConn)

	// Group routes that require authentication
	authGroup := router.Group("/api", middleware.AuthMiddleware(authConn))
	{
		authGroup.POST("/addWaitlist", wt.AddToWaitlist())
		authGroup.GET("/getWaitlist", wt.GetWaitList())
		authGroup.DELETE("/deleteWaitlist/:email", wt.DeleteFromWaitlist())
	}

	// Routes that do not require authentication
	router.POST("/api/signin", wt.Signin())
	router.POST("/api/create", wt.CreateAdmin())
}
