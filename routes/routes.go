package routes

import (
	"waitlist/controllers"
	"waitlist/db"
	"waitlist/lib/emailclient/postmark"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	email := postmark.New()
	wt := controllers.NewWaitlist(db.ConnectDatabase(), email)

	router.POST("/api/addWaitlist", wt.AddToWaitlist())
	router.GET("/api/getWaitlist", wt.GetWaitList())
	router.DELETE("/api/deleteWaitlist/:email", wt.DeleteFromWaitlist())
}
