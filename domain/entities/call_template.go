package entities

import (
	"time"
)

// CallTemplateDataModel represents a call template in the database.
type CallTemplateDataModel struct {
	ID              string    `bson:"id,omitempty" json:"id"`
	TemplateID      string    `bson:"template_id,omitempty" json:"template_id"`
	OrgName         string    `bson:"org_name,omitempty" json:"org_name"`
	Message         string    `bson:"message,omitempty" json:"message"`
	IsSystemDefault bool      `bson:"is_system_default,omitempty" json:"is_system_default"`
	UserID          string    `bson:"user_id,omitempty" json:"user_id"`
	WorkspaceID     string    `bson:"workspace_id,omitempty" json:"workspace_id"`
	CreatedAt       time.Time `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}