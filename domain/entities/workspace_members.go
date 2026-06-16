package entities

import (
	"time"
)

type WorkspaceMemberDataModel struct {
	ID          string    `json:"id" bson:"id,omitempty"`
	WorkspaceID string    `json:"workspace_id" bson:"workspace_id,omitempty"`
	UserID      string    `json:"user_id" bson:"user_id,omitempty"`
	Role        string    `json:"role" bson:"role,omitempty"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at,omitempty"`
}