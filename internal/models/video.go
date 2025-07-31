package models

import (
	"gorm.io/gorm"
	"time"
)

// Video represents a video in the system
type Video struct {
	ID           string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TenantID     string         `json:"tenant_id" gorm:"type:varchar(36);not null;index:idx_tenant_user"`
	UserID       string         `json:"user_id" gorm:"type:varchar(36);not null;index:idx_tenant_user"`
	Title        string         `json:"title" gorm:"type:varchar(255);not null"`
	Description  string         `json:"description" gorm:"type:text"`
	FileName     string         `json:"file_name" gorm:"type:varchar(255);not null"`
	FilePath     string         `json:"file_path" gorm:"type:varchar(500)"`
	FileSize     int64          `json:"file_size" gorm:"not null"`
	Duration     int            `json:"duration" gorm:"default:0"` // in seconds
	Format       string         `json:"format" gorm:"type:varchar(50)"`
	Resolution   string         `json:"resolution" gorm:"type:varchar(50)"`
	Status       string         `json:"status" gorm:"type:varchar(50);not null;index:idx_status;default:'uploading'"`
	Metadata     string         `json:"metadata" gorm:"type:json"` // JSON string
	ThumbnailURL string         `json:"thumbnail_url" gorm:"type:varchar(500)"`
	S3Key        string         `json:"s3_key" gorm:"type:varchar(500)"`
	S3Bucket     string         `json:"s3_bucket" gorm:"type:varchar(255)"`
	Tags         string         `json:"tags" gorm:"type:json"` // JSON array as string
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// VideoStatus defines video processing statuses
type VideoStatus string

const (
	StatusUploading  VideoStatus = "uploading"
	StatusProcessing VideoStatus = "processing"
	StatusReady      VideoStatus = "ready"
	StatusFailed     VideoStatus = "failed"
	StatusArchived   VideoStatus = "archived"
)

// CreateVideoRequest represents the request to create a new video
type CreateVideoRequest struct {
	Title       string   `json:"title" validate:"required,max=255"`
	Description string   `json:"description" validate:"max=1000"`
	FileName    string   `json:"file_name" validate:"required"`
	FileSize    int64    `json:"file_size" validate:"required,min=1"`
	Format      string   `json:"format" validate:"required"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateVideoRequest represents the request to update a video
type UpdateVideoRequest struct {
	Title       *string  `json:"title,omitempty" validate:"omitempty,max=255"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	Tags        []string `json:"tags,omitempty"`
}

// VideoRepository defines the interface for video operations
type VideoRepository interface {
	Create(video *Video) error
	GetByID(tenantID, id string) (*Video, error)
	GetByUserID(tenantID, userID string, limit, offset int) ([]*Video, error)
	Update(video *Video) error
	Delete(tenantID, id string) error
	List(tenantID string, limit, offset int) ([]*Video, error)
	UpdateStatus(tenantID, id string, status VideoStatus) error
	GetByStatus(tenantID string, status VideoStatus, limit, offset int) ([]*Video, error)
}

// VideoService handles business logic for videos
type VideoService struct {
	repo VideoRepository
}

// NewVideoService creates a new video service
func NewVideoService(repo VideoRepository) *VideoService {
	return &VideoService{repo: repo}
}

// CreateVideo creates a new video
func (s *VideoService) CreateVideo(tenantID, userID string, req *CreateVideoRequest) (*Video, error) {
	video := &Video{
		TenantID:    tenantID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		Format:      req.Format,
		Status:      string(StatusUploading),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if len(req.Tags) > 0 {
		// Convert tags to JSON string (simplified)
		video.Tags = convertTagsToJSON(req.Tags)
	}

	if err := s.repo.Create(video); err != nil {
		return nil, err
	}

	return video, nil
}

// GetVideo retrieves a video by ID
func (s *VideoService) GetVideo(tenantID, id string) (*Video, error) {
	return s.repo.GetByID(tenantID, id)
}

// GetUserVideos retrieves videos for a specific user
func (s *VideoService) GetUserVideos(tenantID, userID string, limit, offset int) ([]*Video, error) {
	return s.repo.GetByUserID(tenantID, userID, limit, offset)
}

// UpdateVideo updates an existing video
func (s *VideoService) UpdateVideo(tenantID, id string, req *UpdateVideoRequest) (*Video, error) {
	video, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		video.Title = *req.Title
	}
	if req.Description != nil {
		video.Description = *req.Description
	}
	if len(req.Tags) > 0 {
		video.Tags = convertTagsToJSON(req.Tags)
	}

	video.UpdatedAt = time.Now()

	if err := s.repo.Update(video); err != nil {
		return nil, err
	}

	return video, nil
}

// DeleteVideo soft deletes a video
func (s *VideoService) DeleteVideo(tenantID, id string) error {
	return s.repo.Delete(tenantID, id)
}

// ListVideos retrieves a list of videos with pagination
func (s *VideoService) ListVideos(tenantID string, limit, offset int) ([]*Video, error) {
	return s.repo.List(tenantID, limit, offset)
}

// UpdateVideoStatus updates the processing status of a video
func (s *VideoService) UpdateVideoStatus(tenantID, id string, status VideoStatus) error {
	return s.repo.UpdateStatus(tenantID, id, status)
}

// GetVideosByStatus retrieves videos by status
func (s *VideoService) GetVideosByStatus(tenantID string, status VideoStatus, limit, offset int) ([]*Video, error) {
	return s.repo.GetByStatus(tenantID, status, limit, offset)
}

// SetProcessingComplete marks a video as ready and updates metadata
func (s *VideoService) SetProcessingComplete(tenantID, id string, duration int, resolution, thumbnailURL, s3Key, s3Bucket string) error {
	video, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return err
	}

	video.Duration = duration
	video.Resolution = resolution
	video.ThumbnailURL = thumbnailURL
	video.S3Key = s3Key
	video.S3Bucket = s3Bucket
	video.Status = string(StatusReady)
	video.UpdatedAt = time.Now()

	return s.repo.Update(video)
}

// SetProcessingFailed marks a video as failed
func (s *VideoService) SetProcessingFailed(tenantID, id string) error {
	return s.repo.UpdateStatus(tenantID, id, StatusFailed)
}

// IsReady checks if the video is ready for publishing
func (v *Video) IsReady() bool {
	return v.Status == string(StatusReady) && !v.DeletedAt.Valid
}

// IsProcessing checks if the video is currently being processed
func (v *Video) IsProcessing() bool {
	return v.Status == string(StatusProcessing) || v.Status == string(StatusUploading)
}

// GetTags converts the JSON tags string back to a slice
func (v *Video) GetTags() []string {
	// Simplified implementation - in real app, use proper JSON unmarshaling
	if v.Tags == "" {
		return []string{}
	}
	// This is a placeholder - implement proper JSON parsing
	return []string{}
}

// convertTagsToJSON converts a slice of tags to JSON string
func convertTagsToJSON(tags []string) string {
	// Simplified implementation - in real app, use proper JSON marshaling
	if len(tags) == 0 {
		return ""
	}
	// This is a placeholder - implement proper JSON marshaling
	return "[\"" + tags[0] + "\"]"
}
