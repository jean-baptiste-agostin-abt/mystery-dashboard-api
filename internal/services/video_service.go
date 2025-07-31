package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibe0123/mysteryfactory/internal/models"
	"github.com/jibe0123/mysteryfactory/pkg/logger"
)

// videoService implements the VideoService interface
type videoService struct {
	repo   models.VideoRepository
	logger *logger.Logger
}

// NewVideoService creates a new video service instance
func NewVideoService(repo models.VideoRepository, logger *logger.Logger) VideoService {
	return &videoService{
		repo:   repo,
		logger: logger,
	}
}

// CreateVideo creates a new video
func (s *videoService) CreateVideo(ctx context.Context, tenantID, userID string, req *models.CreateVideoRequest) (*models.Video, error) {
	s.logger.Info("Creating video", "tenant_id", tenantID, "user_id", userID, "title", req.Title)

	// Validate request
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.FileName == "" {
		return nil, fmt.Errorf("file name is required")
	}

	// Create video entity
	video := &models.Video{
		ID:          generateID(),
		TenantID:    tenantID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		Format:      req.Format,
		Status:      string(models.StatusUploading),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Convert tags to JSON if provided
	if len(req.Tags) > 0 {
		video.Tags = convertTagsToJSON(req.Tags)
	}

	// Save to repository
	if err := s.repo.Create(video); err != nil {
		s.logger.Error("Failed to create video", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to create video: %w", err)
	}

	s.logger.Info("Video created successfully", "video_id", video.ID, "tenant_id", tenantID)
	return video, nil
}

// GetVideo retrieves a video by ID
func (s *videoService) GetVideo(ctx context.Context, tenantID, videoID string) (*models.Video, error) {
	s.logger.Debug("Getting video", "video_id", videoID, "tenant_id", tenantID)

	video, err := s.repo.GetByID(tenantID, videoID)
	if err != nil {
		s.logger.Error("Failed to get video", "error", err, "video_id", videoID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	return video, nil
}

// UpdateVideo updates an existing video
func (s *videoService) UpdateVideo(ctx context.Context, tenantID, videoID string, req *models.UpdateVideoRequest) (*models.Video, error) {
	s.logger.Info("Updating video", "video_id", videoID, "tenant_id", tenantID)

	// Get existing video
	video, err := s.repo.GetByID(tenantID, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	// Update fields
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

	// Save changes
	if err := s.repo.Update(video); err != nil {
		s.logger.Error("Failed to update video", "error", err, "video_id", videoID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update video: %w", err)
	}

	s.logger.Info("Video updated successfully", "video_id", videoID, "tenant_id", tenantID)
	return video, nil
}

// DeleteVideo deletes a video
func (s *videoService) DeleteVideo(ctx context.Context, tenantID, videoID string) error {
	s.logger.Info("Deleting video", "video_id", videoID, "tenant_id", tenantID)

	if err := s.repo.Delete(tenantID, videoID); err != nil {
		s.logger.Error("Failed to delete video", "error", err, "video_id", videoID, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete video: %w", err)
	}

	s.logger.Info("Video deleted successfully", "video_id", videoID, "tenant_id", tenantID)
	return nil
}

// ListVideos lists videos for a tenant
func (s *videoService) ListVideos(ctx context.Context, tenantID string, limit, offset int) ([]*models.Video, error) {
	s.logger.Debug("Listing videos", "tenant_id", tenantID, "limit", limit, "offset", offset)

	videos, err := s.repo.List(tenantID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list videos", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to list videos: %w", err)
	}

	return videos, nil
}

// GetUserVideos retrieves videos for a specific user
func (s *videoService) GetUserVideos(ctx context.Context, tenantID, userID string, limit, offset int) ([]*models.Video, error) {
	s.logger.Debug("Getting user videos", "tenant_id", tenantID, "user_id", userID, "limit", limit, "offset", offset)

	videos, err := s.repo.GetByUserID(tenantID, userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get user videos", "error", err, "tenant_id", tenantID, "user_id", userID)
		return nil, fmt.Errorf("failed to get user videos: %w", err)
	}

	return videos, nil
}

// UploadVideo handles video file upload
func (s *videoService) UploadVideo(ctx context.Context, tenantID, videoID string, fileData []byte, filename string) error {
	s.logger.Info("Uploading video file", "video_id", videoID, "tenant_id", tenantID, "filename", filename)

	// Get video to ensure it exists
	video, err := s.repo.GetByID(tenantID, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	// TODO: Implement S3 upload logic here
	// For now, just update the status
	video.Status = string(models.StatusProcessing)
	video.UpdatedAt = time.Now()

	if err := s.repo.Update(video); err != nil {
		s.logger.Error("Failed to update video status", "error", err, "video_id", videoID, "tenant_id", tenantID)
		return fmt.Errorf("failed to update video status: %w", err)
	}

	s.logger.Info("Video file uploaded successfully", "video_id", videoID, "tenant_id", tenantID)
	return nil
}

// UpdateVideoStatus updates the status of a video
func (s *videoService) UpdateVideoStatus(ctx context.Context, tenantID, videoID string, status models.VideoStatus) error {
	s.logger.Info("Updating video status", "video_id", videoID, "tenant_id", tenantID, "status", status)

	if err := s.repo.UpdateStatus(tenantID, videoID, status); err != nil {
		s.logger.Error("Failed to update video status", "error", err, "video_id", videoID, "tenant_id", tenantID)
		return fmt.Errorf("failed to update video status: %w", err)
	}

	s.logger.Info("Video status updated successfully", "video_id", videoID, "tenant_id", tenantID, "status", status)
	return nil
}

// GetVideosByStatus retrieves videos by status
func (s *videoService) GetVideosByStatus(ctx context.Context, tenantID string, status models.VideoStatus, limit, offset int) ([]*models.Video, error) {
	s.logger.Debug("Getting videos by status", "tenant_id", tenantID, "status", status, "limit", limit, "offset", offset)

	videos, err := s.repo.GetByStatus(tenantID, status, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get videos by status", "error", err, "tenant_id", tenantID, "status", status)
		return nil, fmt.Errorf("failed to get videos by status: %w", err)
	}

	return videos, nil
}

// SetProcessingComplete marks video processing as complete
func (s *videoService) SetProcessingComplete(ctx context.Context, tenantID, videoID string, duration int, resolution, thumbnailURL, s3Key, s3Bucket string) error {
	s.logger.Info("Setting video processing complete", "video_id", videoID, "tenant_id", tenantID)

	video, err := s.repo.GetByID(tenantID, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	video.Status = string(models.StatusReady)
	video.Duration = duration
	video.Resolution = resolution
	video.ThumbnailURL = thumbnailURL
	video.S3Key = s3Key
	video.S3Bucket = s3Bucket
	video.UpdatedAt = time.Now()

	if err := s.repo.Update(video); err != nil {
		s.logger.Error("Failed to update video processing status", "error", err, "video_id", videoID, "tenant_id", tenantID)
		return fmt.Errorf("failed to update video processing status: %w", err)
	}

	s.logger.Info("Video processing completed successfully", "video_id", videoID, "tenant_id", tenantID)
	return nil
}

// SetProcessingFailed marks video processing as failed
func (s *videoService) SetProcessingFailed(ctx context.Context, tenantID, videoID string) error {
	s.logger.Info("Setting video processing failed", "video_id", videoID, "tenant_id", tenantID)

	video, err := s.repo.GetByID(tenantID, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	video.Status = string(models.StatusFailed)
	video.UpdatedAt = time.Now()

	if err := s.repo.Update(video); err != nil {
		s.logger.Error("Failed to update video processing status", "error", err, "video_id", videoID, "tenant_id", tenantID)
		return fmt.Errorf("failed to update video processing status: %w", err)
	}

	s.logger.Info("Video processing marked as failed", "video_id", videoID, "tenant_id", tenantID)
	return nil
}

// PublishVideo publishes a video to specified platforms
func (s *videoService) PublishVideo(ctx context.Context, tenantID, videoID string, platforms []string) error {
	s.logger.Info("Publishing video", "video_id", videoID, "tenant_id", tenantID, "platforms", platforms)

	// Get video to ensure it exists and is ready
	video, err := s.repo.GetByID(tenantID, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	if video.Status != string(models.StatusReady) {
		return fmt.Errorf("video is not ready for publishing, current status: %s", video.Status)
	}

	// TODO: Implement publication job creation logic
	// For now, just log the action
	s.logger.Info("Video publication initiated", "video_id", videoID, "tenant_id", tenantID, "platforms", platforms)
	return nil
}

// GetVideoPublications retrieves publication jobs for a video
func (s *videoService) GetVideoPublications(ctx context.Context, tenantID, videoID string) ([]*models.PublicationJob, error) {
	s.logger.Debug("Getting video publications", "video_id", videoID, "tenant_id", tenantID)

	// TODO: Implement publication job retrieval
	// For now, return empty slice
	return []*models.PublicationJob{}, nil
}

// UpdatePublication updates a publication job status
func (s *videoService) UpdatePublication(ctx context.Context, tenantID, publicationID string, status string) error {
	s.logger.Info("Updating publication", "publication_id", publicationID, "tenant_id", tenantID, "status", status)

	// TODO: Implement publication job update logic
	return nil
}

// CancelPublication cancels a publication job
func (s *videoService) CancelPublication(ctx context.Context, tenantID, publicationID string) error {
	s.logger.Info("Cancelling publication", "publication_id", publicationID, "tenant_id", tenantID)

	// TODO: Implement publication job cancellation logic
	return nil
}

// convertTagsToJSON converts a slice of tags to JSON string
func convertTagsToJSON(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	jsonBytes, err := json.Marshal(tags)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

// generateID generates a unique ID for entities
// TODO: Replace with proper UUID generation
func generateID() string {
	return fmt.Sprintf("video_%d", time.Now().UnixNano())
}
