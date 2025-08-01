package models

import (
	"time"

	"gorm.io/gorm"
)

// Workspace represents a user workspace or channel
// allowing a user to manage multiple channels.
type Workspace struct {
	ID       string `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TenantID string `json:"tenant_id" gorm:"type:varchar(36);not null;index"`
	UserID   string `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Name     string `json:"name" gorm:"type:varchar(255);not null"`

	// OAuth and API credentials for partner platforms
	CredentialsPath       string `json:"credentials_path" gorm:"type:varchar(255)"`
	TokenDir              string `json:"token_dir" gorm:"type:varchar(255)"`
	TikTokAppID           string `json:"tiktok_app_id" gorm:"type:varchar(255)"`
	TikTokSecret          string `json:"tiktok_secret" gorm:"type:varchar(255)"`
	RedirectURI           string `json:"redirect_uri" gorm:"type:varchar(500)"`
	OAuthCode             string `json:"oauth_code" gorm:"-"`
	InstagramUserID       string `json:"instagram_user_id" gorm:"type:varchar(255)"`
	InstagramAccessToken  string `json:"instagram_access_token" gorm:"type:varchar(500)"`
	FacebookPageID        string `json:"facebook_page_id" gorm:"type:varchar(255)"`
	FacebookPageToken     string `json:"facebook_page_token" gorm:"type:varchar(500)"`
	TwitterConsumerKey    string `json:"twitter_consumer_key" gorm:"type:varchar(255)"`
	TwitterConsumerSecret string `json:"twitter_consumer_secret" gorm:"type:varchar(255)"`
	TwitterAccessToken    string `json:"twitter_access_token" gorm:"type:varchar(255)"`
	TwitterAccessSecret   string `json:"twitter_access_secret" gorm:"type:varchar(255)"`
	SnapchatAccessToken   string `json:"snapchat_access_token" gorm:"type:varchar(500)"`
	SnapchatProfileID     string `json:"snapchat_profile_id" gorm:"type:varchar(255)"`

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
