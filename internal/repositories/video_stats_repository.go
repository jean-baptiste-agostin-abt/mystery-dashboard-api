package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

type videoStatsRepository struct {
	db *gorm.DB
}

// NewVideoStatsRepository creates a new repository.
func NewVideoStatsRepository(db *gorm.DB) models.VideoStatsRepository {
	return &videoStatsRepository{db: db}
}

func (r *videoStatsRepository) Create(stats *models.VideoStats) error {
	if stats.ID == "" {
		stats.ID = uuid.New().String()
	}
	return r.db.Create(stats).Error
}

func (r *videoStatsRepository) GetByID(tenantID, id string) (*models.VideoStats, error) {
	var s models.VideoStats
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrNotFound
	}
	return &s, err
}

func (r *videoStatsRepository) GetByVideoID(tenantID, videoID string) ([]*models.VideoStats, error) {
	var stats []*models.VideoStats
	err := r.db.Where("tenant_id = ? AND video_id = ?", tenantID, videoID).Find(&stats).Error
	return stats, err
}

func (r *videoStatsRepository) GetByVideoAndPlatform(tenantID, videoID, platform string) (*models.VideoStats, error) {
	var s models.VideoStats
	err := r.db.Where("tenant_id = ? AND video_id = ? AND platform = ?", tenantID, videoID, platform).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrNotFound
	}
	return &s, err
}

func (r *videoStatsRepository) Update(stats *models.VideoStats) error {
	return r.db.Save(stats).Error
}

func (r *videoStatsRepository) Delete(tenantID, id string) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.VideoStats{}).Error
}

func (r *videoStatsRepository) List(tenantID string, limit, offset int) ([]*models.VideoStats, error) {
	var stats []*models.VideoStats
	err := r.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Find(&stats).Error
	return stats, err
}

func (r *videoStatsRepository) GetByPlatform(tenantID, platform string, limit, offset int) ([]*models.VideoStats, error) {
	var stats []*models.VideoStats
	err := r.db.Where("tenant_id = ? AND platform = ?", tenantID, platform).Limit(limit).Offset(offset).Find(&stats).Error
	return stats, err
}

func (r *videoStatsRepository) GetTopPerforming(tenantID string, metric string, limit int) ([]*models.VideoStats, error) {
	var stats []*models.VideoStats
	order := metric + " DESC"
	err := r.db.Where("tenant_id = ?", tenantID).Order(order).Limit(limit).Find(&stats).Error
	return stats, err
}

func (r *videoStatsRepository) GetAggregatedStats(tenantID, videoID string) (*models.StatsAggregation, error) {
	var agg models.StatsAggregation
	err := r.db.Model(&models.VideoStats{}).
		Select("video_id, SUM(views) as total_views, SUM(likes) as total_likes, SUM(comments) as total_comments, SUM(shares) as total_shares, SUM(revenue) as total_revenue, COUNT(platform) as platform_count").
		Where("tenant_id = ? AND video_id = ?", tenantID, videoID).
		Group("video_id").
		Scan(&agg).Error
	return &agg, err
}

func (r *videoStatsRepository) CreateSnapshot(snapshot *models.VideoStatsSnapshot) error {
	if snapshot.ID == "" {
		snapshot.ID = uuid.New().String()
	}
	return r.db.Create(snapshot).Error
}

func (r *videoStatsRepository) GetSnapshots(statsID string, limit int) ([]*models.VideoStatsSnapshot, error) {
	var snaps []*models.VideoStatsSnapshot
	err := r.db.Where("stats_id = ?", statsID).Order("created_at DESC").Limit(limit).Find(&snaps).Error
	return snaps, err
}

func (r *videoStatsRepository) GetStatsNeedingSync(olderThan time.Time, limit int) ([]*models.VideoStats, error) {
	var stats []*models.VideoStats
	err := r.db.Where("last_sync_at <= ?", olderThan).Limit(limit).Find(&stats).Error
	return stats, err
}
