package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"waitlist/lib/emailclient"
	"waitlist/middleware"
	"waitlist/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Waitlist struct {
	db          *mongo.Database
	emailclient emailclient.EmailClient
	auth        *middleware.AuthConn
}

const (
	// email templates
	WaitlistAlias = "waitlist-signup"
)

func NewWaitlist(db *mongo.Database, email emailclient.EmailClient, auth *middleware.AuthConn) *Waitlist {
	return &Waitlist{
		db:          db,
		emailclient: email,
		auth:        auth,
	}
}

func (w *Waitlist) AddToWaitlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		var waitlistEntry models.WaitlistEntry
		collection := w.db.Collection("waitlist")

		if err := c.BindJSON(&waitlistEntry); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error": "Unable to bind waitlist"})
			return
		}

		filter := bson.M{"email": waitlistEntry.Email}
		result := collection.FindOne(ctx, filter)

		entry := models.WaitlistEntry{}
		err := result.Decode(&entry)
		if entry.Email != "" {
			c.AbortWithStatusJSON(http.StatusAlreadyReported, gin.H{"message": "Email already added to waitlist"})
			return
		}

		if err == mongo.ErrNoDocuments {
			waitlistEntry.Timestamp = time.Now().Unix()
			_, err := collection.InsertOne(ctx, waitlistEntry)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, "Database error")
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "User added to waitlist"})
		} else if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, "Database error")
			return
		}
		err = w.sendMsg(waitlistEntry.Email, "waitlist-signup", WaitlistAlias)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": "Unable to send email"})
			return
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
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var entry models.WaitlistEntry
			if err := cursor.Decode(&entry); err != nil {
				log.Println("MongoDb decode error", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": "error decoding document"})
				return
			}
			waitlist = append(waitlist, entry)
		}

		c.JSON(http.StatusOK, waitlist)
	}
}

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

func (w *Waitlist) Signin() gin.HandlerFunc {
	return func(c *gin.Context) {
		collection := w.db.Collection("admin")
		ctx := context.Background()

		user := models.SiginDetails{}

		if err := c.BindJSON(&user); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unable to decode Json", "msg": err.Error()})
			return
		}

		filter := bson.M{"email": user.Email}
		result := collection.FindOne(ctx, filter)

		userDetails := models.SiginDetails{}

		if err := result.Decode(&userDetails); err != nil {
			if err == mongo.ErrNoDocuments {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "could not find record for email"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error while getting record for email", "msg": err.Error()})
			return
		}

		valid := comparePasswords(user.Password, userDetails.Password)
		if !valid {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "incorrect password"})
			return
		}

		token, err := w.auth.GenerateJWT(userDetails.Email)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unable to generate token"})
			return
		}

		c.Header("Authorization", token)

		c.JSON(http.StatusAccepted, gin.H{"message": "login successful"})
	}
}

func (w *Waitlist) CreateAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {

		// check if that email already exists
		collection := w.db.Collection("admin")
		ctx := context.Background()

		user := models.SiginDetails{}
		if err := c.BindJSON(&user); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unable to decode Json payload", "message": err.Error()})
			return
		}

		hashPasswrd, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password", "message": err.Error()})
			return
		}

		user.Password = string(hashPasswrd)

		_, err = collection.InsertOne(ctx, user)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unable to write to database", "message": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"message": "successfully created admin"})
	}
}

func comparePasswords(inputPasswrd, hashPasswrd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPasswrd), []byte(inputPasswrd))
	return err == nil
}
