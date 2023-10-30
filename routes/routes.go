package routes

import (
	"waitlist/controllers"
	"waitlist/db"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	wt := controllers.NewWaitlist(db.ConnectDatabase())

	router.POST("/api/addWaitlist", wt.AddToWaitlist())
	router.GET("/api/getWaitlist", wt.GetWaitList())
	router.DELETE("/api/deleteWaitlist", wt.DeleteFromWaitlist())
}
