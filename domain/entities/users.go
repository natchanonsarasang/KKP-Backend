package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserDataModel struct {
	UserID      string             `json:"user_id" bson:"user_id,omitempty"`
	WorkspaceID primitive.ObjectID `json:"workspace_id" bson:"workspace_id,omitempty"`
	Username    string             `json:"username" bson:"username,omitempty"`
	Email       string             `json:"email" bson:"email,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at,omitempty"`
	Ip          string             `json:"ip" bson:"ip,omitempty"`
}
