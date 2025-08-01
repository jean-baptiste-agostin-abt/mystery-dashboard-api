package models

import (
	"database/sql"
	"time"
)

// PublicationJob represents a video publication job to a platform
type PublicationJob struct {
	ID          string       `json:"id" db:"id"`
	TenantID    string       `json:"tenant_id" db:"tenant_id"`
	VideoID     string       `json:"video_id" db:"video_id"`
	UserID      string       `json:"user_id" db:"user_id"`
	Platform    string       `json:"platform" db:"platform"`
	Status      string       `json:"status" db:"status"`
	Config      string       `json:"config" db:"config"`             // JSON string with platform-specific config
	ExternalID  string       `json:"external_id" db:"external_id"`   // Platform's video ID
	ExternalURL string       `json:"external_url" db:"external_url"` // Platform's video URL
	ErrorMsg    string       `json:"error_message,omitempty" db:"error_message"`
	RetryCount  int          `json:"retry_count" db:"retry_count"`
	MaxRetries  int          `json:"max_retries" db:"max_retries"`
	ScheduledAt sql.NullTime `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt   sql.NullTime `json:"started_at,omitempty" db:"started_at"`
	CompletedAt sql.NullTime `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt   sql.NullTime `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PublicationStatus defines publication job statuses
type PublicationStatus string

const (
	PublicationPending    PublicationStatus = "pending"
	PublicationScheduled  PublicationStatus = "scheduled"
	PublicationProcessing PublicationStatus = "processing"
	PublicationCompleted  PublicationStatus = "completed"
	PublicationFailed     PublicationStatus = "failed"
	PublicationCancelled  PublicationStatus = "cancelled"
)

// Platform defines supported platforms
type Platform string

const (
	PlatformYouTube   Platform = "youtube"
	PlatformTikTok    Platform = "tiktok"
	PlatformInstagram Platform = "instagram"
	PlatformFacebook  Platform = "facebook"
	PlatformTwitter   Platform = "twitter"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformSnapchat  Platform = "snapchat"
)

// CreatePublicationJobRequest represents the request to create a publication job
type CreatePublicationJobRequest struct {
	VideoID     string                 `json:"video_id" validate:"required"`
	Platform    string                 `json:"platform" validate:"required,oneof=youtube tiktok instagram facebook twitter linkedin snapchat"`
	Config      map[string]interface{} `json:"config,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	MaxRetries  int                    `json:"max_retries,omitempty"`
}

// UpdatePublicationJobRequest represents the request to update a publication job
type UpdatePublicationJobRequest struct {
	Status      *string                `json:"status,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

// PublicationJobRepository defines the interface for publication job operations
type PublicationJobRepository interface {
	Create(job *PublicationJob) error
	GetByID(tenantID, id string) (*PublicationJob, error)
	GetByVideoID(tenantID, videoID string) ([]*PublicationJob, error)
	GetByStatus(tenantID string, status PublicationStatus, limit, offset int) ([]*PublicationJob, error)
	GetByPlatform(tenantID string, platform Platform, limit, offset int) ([]*PublicationJob, error)
	GetScheduledJobs(before time.Time, limit int) ([]*PublicationJob, error)
	Update(job *PublicationJob) error
	Delete(tenantID, id string) error
	List(tenantID string, limit, offset int) ([]*PublicationJob, error)
	UpdateStatus(tenantID, id string, status PublicationStatus) error
	IncrementRetryCount(tenantID, id string) error
}

// PublicationJobService handles business logic for publication jobs
type PublicationJobService struct {
	repo PublicationJobRepository
}

// NewPublicationJobService creates a new publication job service
func NewPublicationJobService(repo PublicationJobRepository) *PublicationJobService {
	return &PublicationJobService{repo: repo}
}

// CreatePublicationJob creates a new publication job
func (s *PublicationJobService) CreatePublicationJob(tenantID, userID string, req *CreatePublicationJobRequest) (*PublicationJob, error) {
	maxRetries := req.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3 // Default max retries
	}

	job := &PublicationJob{
		TenantID:   tenantID,
		VideoID:    req.VideoID,
		UserID:     userID,
		Platform:   req.Platform,
		Status:     string(PublicationPending),
		MaxRetries: maxRetries,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if req.Config != nil {
		job.Config = convertConfigToJSON(req.Config)
	}

	if req.ScheduledAt != nil {
		job.ScheduledAt = sql.NullTime{Time: *req.ScheduledAt, Valid: true}
		job.Status = string(PublicationScheduled)
	}

	if err := s.repo.Create(job); err != nil {
		return nil, err
	}

	return job, nil
}

// GetPublicationJob retrieves a publication job by ID
func (s *PublicationJobService) GetPublicationJob(tenantID, id string) (*PublicationJob, error) {
	return s.repo.GetByID(tenantID, id)
}

// GetVideoPublicationJobs retrieves all publication jobs for a video
func (s *PublicationJobService) GetVideoPublicationJobs(tenantID, videoID string) ([]*PublicationJob, error) {
	return s.repo.GetByVideoID(tenantID, videoID)
}

// UpdatePublicationJob updates an existing publication job
func (s *PublicationJobService) UpdatePublicationJob(tenantID, id string, req *UpdatePublicationJobRequest) (*PublicationJob, error) {
	job, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if req.Status != nil {
		job.Status = *req.Status
	}
	if req.Config != nil {
		job.Config = convertConfigToJSON(req.Config)
	}
	if req.ScheduledAt != nil {
		job.ScheduledAt = sql.NullTime{Time: *req.ScheduledAt, Valid: true}
	}

	job.UpdatedAt = time.Now()

	if err := s.repo.Update(job); err != nil {
		return nil, err
	}

	return job, nil
}

// DeletePublicationJob soft deletes a publication job
func (s *PublicationJobService) DeletePublicationJob(tenantID, id string) error {
	return s.repo.Delete(tenantID, id)
}

// ListPublicationJobs retrieves a list of publication jobs with pagination
func (s *PublicationJobService) ListPublicationJobs(tenantID string, limit, offset int) ([]*PublicationJob, error) {
	return s.repo.List(tenantID, limit, offset)
}

// GetJobsByStatus retrieves publication jobs by status
func (s *PublicationJobService) GetJobsByStatus(tenantID string, status PublicationStatus, limit, offset int) ([]*PublicationJob, error) {
	return s.repo.GetByStatus(tenantID, status, limit, offset)
}

// GetJobsByPlatform retrieves publication jobs by platform
func (s *PublicationJobService) GetJobsByPlatform(tenantID string, platform Platform, limit, offset int) ([]*PublicationJob, error) {
	return s.repo.GetByPlatform(tenantID, platform, limit, offset)
}

// GetScheduledJobs retrieves jobs scheduled before a specific time
func (s *PublicationJobService) GetScheduledJobs(before time.Time, limit int) ([]*PublicationJob, error) {
	return s.repo.GetScheduledJobs(before, limit)
}

// StartJob marks a job as processing
func (s *PublicationJobService) StartJob(tenantID, id string) error {
	job, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return err
	}

	job.Status = string(PublicationProcessing)
	job.StartedAt = sql.NullTime{Time: time.Now(), Valid: true}
	job.UpdatedAt = time.Now()

	return s.repo.Update(job)
}

// CompleteJob marks a job as completed
func (s *PublicationJobService) CompleteJob(tenantID, id, externalID, externalURL string) error {
	job, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return err
	}

	job.Status = string(PublicationCompleted)
	job.ExternalID = externalID
	job.ExternalURL = externalURL
	job.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	job.UpdatedAt = time.Now()

	return s.repo.Update(job)
}

// FailJob marks a job as failed and increments retry count
func (s *PublicationJobService) FailJob(tenantID, id, errorMsg string) error {
	job, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return err
	}

	job.RetryCount++
	job.ErrorMsg = errorMsg
	job.UpdatedAt = time.Now()

	if job.RetryCount >= job.MaxRetries {
		job.Status = string(PublicationFailed)
	} else {
		job.Status = string(PublicationPending) // Retry
	}

	return s.repo.Update(job)
}

// CancelJob marks a job as cancelled
func (s *PublicationJobService) CancelJob(tenantID, id string) error {
	return s.repo.UpdateStatus(tenantID, id, PublicationCancelled)
}

// IsRetryable checks if the job can be retried
func (j *PublicationJob) IsRetryable() bool {
	return j.RetryCount < j.MaxRetries && j.Status == string(PublicationFailed)
}

// IsScheduled checks if the job is scheduled for future execution
func (j *PublicationJob) IsScheduled() bool {
	return j.Status == string(PublicationScheduled) && j.ScheduledAt.Valid
}

// ShouldExecute checks if a scheduled job should be executed now
func (j *PublicationJob) ShouldExecute() bool {
	if !j.IsScheduled() {
		return false
	}
	return j.ScheduledAt.Time.Before(time.Now())
}

// GetPlatformConfig returns the platform-specific configuration
func (j *PublicationJob) GetPlatformConfig() map[string]interface{} {
	// Simplified implementation - in real app, use proper JSON unmarshaling
	if j.Config == "" {
		return make(map[string]interface{})
	}
	// This is a placeholder - implement proper JSON parsing
	return make(map[string]interface{})
}

// convertConfigToJSON converts a config map to JSON string
func convertConfigToJSON(config map[string]interface{}) string {
	// Simplified implementation - in real app, use proper JSON marshaling
	if len(config) == 0 {
		return ""
	}
	// This is a placeholder - implement proper JSON marshaling
	return "{}"
}
