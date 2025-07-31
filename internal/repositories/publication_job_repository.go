package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

// publicationJobRepository implements models.PublicationJobRepository.
type publicationJobRepository struct {
	db *gorm.DB
}

// NewPublicationJobRepository creates a new repository.
func NewPublicationJobRepository(db *gorm.DB) models.PublicationJobRepository {
	return &publicationJobRepository{db: db}
}

func (r *publicationJobRepository) Create(job *models.PublicationJob) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	return r.db.Create(job).Error
}

func (r *publicationJobRepository) GetByID(tenantID, id string) (*models.PublicationJob, error) {
	var job models.PublicationJob
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrPublicationNotFound
	}
	return &job, err
}

func (r *publicationJobRepository) GetByVideoID(tenantID, videoID string) ([]*models.PublicationJob, error) {
	var jobs []*models.PublicationJob
	err := r.db.Where("tenant_id = ? AND video_id = ?", tenantID, videoID).Find(&jobs).Error
	return jobs, err
}

func (r *publicationJobRepository) GetByStatus(tenantID string, status models.PublicationStatus, limit, offset int) ([]*models.PublicationJob, error) {
	var jobs []*models.PublicationJob
	err := r.db.Where("tenant_id = ? AND status = ?", tenantID, status).Limit(limit).Offset(offset).Find(&jobs).Error
	return jobs, err
}

func (r *publicationJobRepository) GetByPlatform(tenantID string, platform models.Platform, limit, offset int) ([]*models.PublicationJob, error) {
	var jobs []*models.PublicationJob
	err := r.db.Where("tenant_id = ? AND platform = ?", tenantID, platform).Limit(limit).Offset(offset).Find(&jobs).Error
	return jobs, err
}

func (r *publicationJobRepository) GetScheduledJobs(before time.Time, limit int) ([]*models.PublicationJob, error) {
	var jobs []*models.PublicationJob
	err := r.db.Where("status = ? AND scheduled_at <= ?", models.PublicationScheduled, before).Limit(limit).Find(&jobs).Error
	return jobs, err
}

func (r *publicationJobRepository) Update(job *models.PublicationJob) error {
	return r.db.Save(job).Error
}

func (r *publicationJobRepository) Delete(tenantID, id string) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.PublicationJob{}).Error
}

func (r *publicationJobRepository) List(tenantID string, limit, offset int) ([]*models.PublicationJob, error) {
	var jobs []*models.PublicationJob
	err := r.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Find(&jobs).Error
	return jobs, err
}

func (r *publicationJobRepository) UpdateStatus(tenantID, id string, status models.PublicationStatus) error {
	return r.db.Model(&models.PublicationJob{}).Where("tenant_id = ? AND id = ?", tenantID, id).Update("status", status).Error
}

func (r *publicationJobRepository) IncrementRetryCount(tenantID, id string) error {
	return r.db.Model(&models.PublicationJob{}).Where("tenant_id = ? AND id = ?", tenantID, id).UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}
