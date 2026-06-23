package entities

import (
	"time"

	"github.com/google/uuid"
)

// WorkspaceDataModel represents a workspace in the database.
type WorkspaceDataModel struct {
	ID        string    `json:"id" bson:"id,omitempty"`
	Name      string    `json:"name" bson:"name,omitempty"`
	OwnerID   string    `json:"owner_id" bson:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at,omitempty"`
}

// NewWorkspace initializes a new WorkspaceDataModel with a UUIDv4 ID and current timestamps.
func NewWorkspace() WorkspaceDataModel {
	now := time.Now().UTC()
	return WorkspaceDataModel{
		ID:        uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type WorkspaceFilter struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"`
}