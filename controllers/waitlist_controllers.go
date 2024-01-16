package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"waitlist/lib/emailclient"
	"waitlist/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Waitlist struct {
	db          *mongo.Database
	emailclient emailclient.EmailClient
}

const (
	// email templates
	WaitlistAlias = "waitlist-signup"
)

func NewWaitlist(db *mongo.Database, email emailclient.EmailClient) *Waitlist {
	return &Waitlist{
		db:          db,
		emailclient: email,
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

		filter := bson.M{"email": waitlistEntry.Email}
		result := collection.FindOne(ctx, filter)

		entry := models.WaitlistEntry{}
		err := result.Decode(&entry)
		if entry.Email != "" {
			c.AbortWithStatusJSON(http.StatusAlreadyReported, gin.H{"message": "Email already added to waitlist"})
		}

		if err == mongo.ErrNoDocuments {
			waitlistEntry.Timestamp = time.Now().Unix()
			_, err := collection.InsertOne(ctx, waitlistEntry)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, "Database error")
			}
			c.JSON(http.StatusOK, gin.H{"message": "User added to waitlist"})
		} else if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, "Database error")
		}
		err = w.sendMsg(waitlistEntry.Email, "waitlist-signup", WaitlistAlias)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": "Unable to send email"})
		}
		c.JSON(http.StatusOK, gin.H{"message": "Email sent to waitlist"})

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

//// delete email from waitlist
//func (w *Waitlist) DeleteFromWaitlist() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		collection := w.db.Collection("waitlist")
//		ctx := context.Background()
//		var waitlistEntry models.WaitlistEntry
//
//		if err := c.BindJSON(&waitlistEntry); err != nil {
//			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error": "Unable to bind waitlist"})
//		}
//
//		filter := bson.M{"email": waitlistEntry.Email}
//		result := collection.FindOne(ctx, filter)
//
//		entry := models.WaitlistEntry{}
//		err := result.Decode(&entry)
//		if err == mongo.ErrNoDocuments {
//			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Email not found"})
//		} else if err != nil {
//			c.AbortWithStatusJSON(http.StatusInternalServerError, "Database error")
//		}
//
//		_, err = collection.DeleteOne(ctx, filter)
//		if err != nil {
//			c.AbortWithStatusJSON(http.StatusInternalServerError, "Database error")
//		}
//
//		c.JSON(http.StatusOK, gin.H{"message": "Email deleted from waitlist"})
//	}
//}

// Delete email from waitlist using URL parameters
func (w *Waitlist) DeleteFromWaitlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		collection := w.db.Collection("waitlist")
		ctx := context.Background()

		// Get the email parameter from the URL
		email := c.Param("email")

		// Check if the email parameter is empty
		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Email parameter is missing"})
			return
		}

		filter := bson.M{"email": email}
		result := collection.FindOne(ctx, filter)

		entry := models.WaitlistEntry{}
		err := result.Decode(&entry)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "Email not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
			return
		}

		_, err = collection.DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email deleted from waitlist"})
	}
}

// send email to waitlist

func (w *Waitlist) sendMsg(Email string, title string, templateID string) error {

	// Creating Message
	message := models.Message{
		Target:     Email,
		Type:       "email",
		Title:      title,
		TemplateID: templateID,
		DataMap:    map[string]string{},
	}
	message.DataMap["Email"] = Email

	// send message
	fmt.Println("about send email")
	if err := w.emailclient.Send(&message); err != nil {
		return err
	}
	fmt.Println("email sent")
	return nil
}
