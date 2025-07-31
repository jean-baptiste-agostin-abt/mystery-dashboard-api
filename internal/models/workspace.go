package models

import (
	"time"

	"gorm.io/gorm"
)

// Workspace represents a user workspace or channel
// allowing a user to manage multiple channels.
type Workspace struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TenantID  string         `json:"tenant_id" gorm:"type:varchar(36);not null;index"`
	UserID    string         `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Name      string         `json:"name" gorm:"type:varchar(255);not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// WorkspaceRepository defines data access methods for workspaces.
type WorkspaceRepository interface {
	Create(workspace *Workspace) error
	GetByID(tenantID, id string) (*Workspace, error)
	ListByUser(tenantID, userID string) ([]*Workspace, error)
	Update(workspace *Workspace) error
	Delete(tenantID, id string) error
}
