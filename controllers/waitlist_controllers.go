package controllers

import (
	"context"
	"log"
	"net/http"
	"waitlist/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Waitlist struct {
	db *mongo.Database
}

func NewWaitlist(db *mongo.Database) *Waitlist {
	return &Waitlist{
		db: db,
	}
}

func (w *Waitlist) AddToWaitlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		var waitlistEntry models.WaitlistEntry
		collection := w.db.Collection("waitlist")

		if err := c.BindJSON(&waitlistEntry); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error": "Unable to bind waitlist"})
		}

		_, err := collection.InsertOne(ctx, waitlistEntry)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, "Database error")
		}

		c.JSON(http.StatusOK, gin.H{"message": "User added to waitlist"})
	}
}

func (w *Waitlist) GetWaitList() gin.HandlerFunc {
	return func(c *gin.Context) {
		collection := w.db.Collection("waitlist")
		waitlist := []models.WaitlistEntry{}
		ctx := context.Background()

		cursor, err := collection.Find(ctx, bson.D{})
		if err != nil {
			log.Println("MongoDv find error:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": "error occured while fetching records"})
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var entry models.WaitlistEntry
			if err := cursor.Decode(&entry); err != nil {
				log.Println("MongoDb decode error", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": "error decoding document"})
			}
			waitlist = append(waitlist, entry)
		}

		c.JSON(http.StatusOK, waitlist)
	}
}
