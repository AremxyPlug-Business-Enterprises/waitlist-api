package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type WaitlistEntry struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	Timestamp int64              `json:"timestamp" bson:"timestamp"`
}

type SiginDetails struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
