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
		authGroup.GET("/getWaitlist", wt.GetWaitList())
		authGroup.DELETE("/deleteWaitlist/:email", wt.DeleteFromWaitlist())
	}

	router.POST("/api/addWaitlist", wt.AddToWaitlist())
	router.POST("/api/signin", wt.Signin())
	router.POST("/api/create", wt.CreateAdmin())
}
