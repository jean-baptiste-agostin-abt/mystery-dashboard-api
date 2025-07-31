package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

// videoRepository implements models.VideoRepository.
type videoRepository struct {
	db *gorm.DB
}

// NewVideoRepository creates a new repository instance.
func NewVideoRepository(db *gorm.DB) models.VideoRepository {
	return &videoRepository{db: db}
}

func (r *videoRepository) Create(video *models.Video) error {
	if video.ID == "" {
		video.ID = uuid.New().String()
	}
	return r.db.Create(video).Error
}

func (r *videoRepository) GetByID(tenantID, id string) (*models.Video, error) {
	var v models.Video
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrVideoNotFound
	}
	return &v, err
}

func (r *videoRepository) GetByUserID(tenantID, userID string, limit, offset int) ([]*models.Video, error) {
	var videos []*models.Video
	err := r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Limit(limit).Offset(offset).Find(&videos).Error
	return videos, err
}

func (r *videoRepository) Update(video *models.Video) error {
	return r.db.Save(video).Error
}

func (r *videoRepository) Delete(tenantID, id string) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.Video{}).Error
}

func (r *videoRepository) List(tenantID string, limit, offset int) ([]*models.Video, error) {
	var videos []*models.Video
	err := r.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Find(&videos).Error
	return videos, err
}

func (r *videoRepository) UpdateStatus(tenantID, id string, status models.VideoStatus) error {
	return r.db.Model(&models.Video{}).Where("tenant_id = ? AND id = ?", tenantID, id).Update("status", status).Error
}

func (r *videoRepository) GetByStatus(tenantID string, status models.VideoStatus, limit, offset int) ([]*models.Video, error) {
	var videos []*models.Video
	err := r.db.Where("tenant_id = ? AND status = ?", tenantID, status).Limit(limit).Offset(offset).Find(&videos).Error
	return videos, err
}
