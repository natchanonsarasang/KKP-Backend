package entities

import "time"

type CallTokenDataModel struct {
	ID        string    `bson:"id,omitempty" json:"id"`
	UserID    string    `bson:"user_id,omitempty" json:"user_id"`
	Token     string    `bson:"token,omitempty" json:"token"`
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}
