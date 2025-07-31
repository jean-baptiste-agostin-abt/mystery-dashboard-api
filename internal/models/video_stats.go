package models

import (
	"time"
	"gorm.io/gorm"
)

// VideoStats represents analytics data for a video
type VideoStats struct {
	ID             string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TenantID       string         `json:"tenant_id" gorm:"type:varchar(36);not null;index:idx_tenant_video"`
	VideoID        string         `json:"video_id" gorm:"type:varchar(36);not null;index:idx_tenant_video;index:idx_video_platform"`
	Platform       string         `json:"platform" gorm:"type:varchar(50);not null;index:idx_video_platform"`
	ExternalID     string         `json:"external_id" gorm:"type:varchar(255)"` // Platform's video ID
	Views          int64          `json:"views" gorm:"default:0"`
	Likes          int64          `json:"likes" gorm:"default:0"`
	Dislikes       int64          `json:"dislikes" gorm:"default:0"`
	Comments       int64          `json:"comments" gorm:"default:0"`
	Shares         int64          `json:"shares" gorm:"default:0"`
	Subscribers    int64          `json:"subscribers" gorm:"default:0"` // Gained from this video
	WatchTime      int64          `json:"watch_time" gorm:"default:0"` // Total watch time in seconds
	AvgWatchTime   float64        `json:"avg_watch_time" gorm:"type:decimal(10,2);default:0"`
	ClickThrough   float64        `json:"click_through_rate" gorm:"type:decimal(5,4);default:0"` // CTR percentage
	Engagement     float64        `json:"engagement_rate" gorm:"type:decimal(5,4);default:0"` // Engagement rate percentage
	Revenue        float64        `json:"revenue" gorm:"type:decimal(10,2);default:0"` // Revenue generated
	Impressions    int64          `json:"impressions" gorm:"default:0"`
	Demographics   string         `json:"demographics" gorm:"type:json"` // JSON string with demographic data
	TrafficSources string         `json:"traffic_sources" gorm:"type:json"` // JSON string with traffic source data
	DeviceTypes    string         `json:"device_types" gorm:"type:json"` // JSON string with device type data
	Locations      string         `json:"locations" gorm:"type:json"` // JSON string with geographic data
	LastSyncAt     time.Time      `json:"last_sync_at" gorm:"type:timestamp;not null"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// VideoStatsSnapshot represents a historical snapshot of video stats
type VideoStatsSnapshot struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	StatsID   string    `json:"stats_id" gorm:"type:varchar(36);not null;index"`
	Views     int64     `json:"views" gorm:"default:0"`
	Likes     int64     `json:"likes" gorm:"default:0"`
	Comments  int64     `json:"comments" gorm:"default:0"`
	Shares    int64     `json:"shares" gorm:"default:0"`
	Revenue   float64   `json:"revenue" gorm:"type:decimal(10,2);default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// StatsAggregation represents aggregated stats across platforms
type StatsAggregation struct {
	VideoID        string  `json:"video_id"`
	TotalViews     int64   `json:"total_views"`
	TotalLikes     int64   `json:"total_likes"`
	TotalComments  int64   `json:"total_comments"`
	TotalShares    int64   `json:"total_shares"`
	TotalRevenue   float64 `json:"total_revenue"`
	PlatformCount  int     `json:"platform_count"`
	AvgEngagement  float64 `json:"avg_engagement"`
	TopPlatform    string  `json:"top_platform"`
}

// ROIMetrics represents return on investment calculations
type ROIMetrics struct {
	VideoID           string  `json:"video_id"`
	ProductionCost    float64 `json:"production_cost"`
	PromotionCost     float64 `json:"promotion_cost"`
	TotalInvestment   float64 `json:"total_investment"`
	TotalRevenue      float64 `json:"total_revenue"`
	NetProfit         float64 `json:"net_profit"`
	ROIPercentage     float64 `json:"roi_percentage"`
	RevenuePerView    float64 `json:"revenue_per_view"`
	CostPerView       float64 `json:"cost_per_view"`
	BreakevenViews    int64   `json:"breakeven_views"`
	PaybackPeriodDays int     `json:"payback_period_days"`
}

// EngagementMetrics represents detailed engagement analytics
type EngagementMetrics struct {
	VideoID              string  `json:"video_id"`
	EngagementRate       float64 `json:"engagement_rate"`
	LikeToViewRatio      float64 `json:"like_to_view_ratio"`
	CommentToViewRatio   float64 `json:"comment_to_view_ratio"`
	ShareToViewRatio     float64 `json:"share_to_view_ratio"`
	SubscriberConversion float64 `json:"subscriber_conversion"`
	AverageWatchTime     float64 `json:"average_watch_time"`
	WatchTimePercentage  float64 `json:"watch_time_percentage"`
	ClickThroughRate     float64 `json:"click_through_rate"`
	BounceRate           float64 `json:"bounce_rate"`
	RetentionRate        float64 `json:"retention_rate"`
	ViralityScore        float64 `json:"virality_score"`
}

// VideoStatsRepository defines the interface for video stats operations
type VideoStatsRepository interface {
	Create(stats *VideoStats) error
	GetByID(tenantID, id string) (*VideoStats, error)
	GetByVideoID(tenantID, videoID string) ([]*VideoStats, error)
	GetByVideoAndPlatform(tenantID, videoID, platform string) (*VideoStats, error)
	Update(stats *VideoStats) error
	Delete(tenantID, id string) error
	List(tenantID string, limit, offset int) ([]*VideoStats, error)
	GetByPlatform(tenantID, platform string, limit, offset int) ([]*VideoStats, error)
	GetTopPerforming(tenantID string, metric string, limit int) ([]*VideoStats, error)
	CreateSnapshot(snapshot *VideoStatsSnapshot) error
	GetSnapshots(statsID string, limit int) ([]*VideoStatsSnapshot, error)
	GetAggregatedStats(tenantID, videoID string) (*StatsAggregation, error)
	GetStatsNeedingSync(olderThan time.Time, limit int) ([]*VideoStats, error)
}

// VideoStatsService handles business logic for video statistics
type VideoStatsService struct {
	repo VideoStatsRepository
}

// NewVideoStatsService creates a new video stats service
func NewVideoStatsService(repo VideoStatsRepository) *VideoStatsService {
	return &VideoStatsService{repo: repo}
}

// CreateVideoStats creates new video statistics
func (s *VideoStatsService) CreateVideoStats(tenantID, videoID, platform, externalID string) (*VideoStats, error) {
	stats := &VideoStats{
		TenantID:   tenantID,
		VideoID:    videoID,
		Platform:   platform,
		ExternalID: externalID,
		LastSyncAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.Create(stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// GetVideoStats retrieves video stats by ID
func (s *VideoStatsService) GetVideoStats(tenantID, id string) (*VideoStats, error) {
	return s.repo.GetByID(tenantID, id)
}

// GetVideoStatsByVideo retrieves all stats for a video
func (s *VideoStatsService) GetVideoStatsByVideo(tenantID, videoID string) ([]*VideoStats, error) {
	return s.repo.GetByVideoID(tenantID, videoID)
}

// GetVideoStatsByPlatform retrieves stats for a video on a specific platform
func (s *VideoStatsService) GetVideoStatsByPlatform(tenantID, videoID, platform string) (*VideoStats, error) {
	return s.repo.GetByVideoAndPlatform(tenantID, videoID, platform)
}

// UpdateVideoStats updates existing video statistics
func (s *VideoStatsService) UpdateVideoStats(tenantID, id string, updates map[string]interface{}) (*VideoStats, error) {
	stats, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	// Create snapshot before updating
	snapshot := &VideoStatsSnapshot{
		StatsID:   stats.ID,
		Views:     stats.Views,
		Likes:     stats.Likes,
		Comments:  stats.Comments,
		Shares:    stats.Shares,
		Revenue:   stats.Revenue,
		CreatedAt: time.Now(),
	}
	s.repo.CreateSnapshot(snapshot)

	// Update stats
	if views, ok := updates["views"].(int64); ok {
		stats.Views = views
	}
	if likes, ok := updates["likes"].(int64); ok {
		stats.Likes = likes
	}
	if dislikes, ok := updates["dislikes"].(int64); ok {
		stats.Dislikes = dislikes
	}
	if comments, ok := updates["comments"].(int64); ok {
		stats.Comments = comments
	}
	if shares, ok := updates["shares"].(int64); ok {
		stats.Shares = shares
	}
	if subscribers, ok := updates["subscribers"].(int64); ok {
		stats.Subscribers = subscribers
	}
	if watchTime, ok := updates["watch_time"].(int64); ok {
		stats.WatchTime = watchTime
	}
	if avgWatchTime, ok := updates["avg_watch_time"].(float64); ok {
		stats.AvgWatchTime = avgWatchTime
	}
	if ctr, ok := updates["click_through_rate"].(float64); ok {
		stats.ClickThrough = ctr
	}
	if engagement, ok := updates["engagement_rate"].(float64); ok {
		stats.Engagement = engagement
	}
	if revenue, ok := updates["revenue"].(float64); ok {
		stats.Revenue = revenue
	}
	if impressions, ok := updates["impressions"].(int64); ok {
		stats.Impressions = impressions
	}

	stats.LastSyncAt = time.Now()
	stats.UpdatedAt = time.Now()

	if err := s.repo.Update(stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// DeleteVideoStats soft deletes video statistics
func (s *VideoStatsService) DeleteVideoStats(tenantID, id string) error {
	return s.repo.Delete(tenantID, id)
}

// ListVideoStats retrieves a list of video stats with pagination
func (s *VideoStatsService) ListVideoStats(tenantID string, limit, offset int) ([]*VideoStats, error) {
	return s.repo.List(tenantID, limit, offset)
}

// GetStatsByPlatform retrieves stats for a specific platform
func (s *VideoStatsService) GetStatsByPlatform(tenantID, platform string, limit, offset int) ([]*VideoStats, error) {
	return s.repo.GetByPlatform(tenantID, platform, limit, offset)
}

// GetTopPerformingVideos retrieves top performing videos by a specific metric
func (s *VideoStatsService) GetTopPerformingVideos(tenantID, metric string, limit int) ([]*VideoStats, error) {
	validMetrics := map[string]bool{
		"views":           true,
		"likes":           true,
		"comments":        true,
		"shares":          true,
		"engagement_rate": true,
		"revenue":         true,
	}

	if !validMetrics[metric] {
		return nil, ErrInvalidInput
	}

	return s.repo.GetTopPerforming(tenantID, metric, limit)
}

// GetAggregatedStats retrieves aggregated statistics for a video across all platforms
func (s *VideoStatsService) GetAggregatedStats(tenantID, videoID string) (*StatsAggregation, error) {
	return s.repo.GetAggregatedStats(tenantID, videoID)
}

// GetStatsHistory retrieves historical snapshots for video stats
func (s *VideoStatsService) GetStatsHistory(tenantID, statsID string, limit int) ([]*VideoStatsSnapshot, error) {
	// Verify the stats belong to the tenant
	_, err := s.repo.GetByID(tenantID, statsID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetSnapshots(statsID, limit)
}

// SyncStatsFromPlatform updates stats with data from external platform
func (s *VideoStatsService) SyncStatsFromPlatform(tenantID, statsID string, platformData map[string]interface{}) error {
	_, err := s.UpdateVideoStats(tenantID, statsID, platformData)
	return err
}

// GetStatsNeedingSync retrieves stats that need to be synced with external platforms
func (s *VideoStatsService) GetStatsNeedingSync(olderThan time.Time, limit int) ([]*VideoStats, error) {
	return s.repo.GetStatsNeedingSync(olderThan, limit)
}

// CalculateEngagementRate calculates engagement rate for video stats
func (s *VideoStats) CalculateEngagementRate() float64 {
	if s.Views == 0 {
		return 0
	}
	
	totalEngagements := s.Likes + s.Comments + s.Shares
	return (float64(totalEngagements) / float64(s.Views)) * 100
}

// CalculateClickThroughRate calculates click-through rate
func (s *VideoStats) CalculateClickThroughRate() float64 {
	if s.Impressions == 0 {
		return 0
	}
	
	return (float64(s.Views) / float64(s.Impressions)) * 100
}

// IsHighPerforming checks if the video is performing well based on engagement
func (s *VideoStats) IsHighPerforming() bool {
	engagementRate := s.CalculateEngagementRate()
	return engagementRate > 5.0 // 5% engagement rate threshold
}

// GetPerformanceScore calculates a composite performance score
func (s *VideoStats) GetPerformanceScore() float64 {
	if s.Views == 0 {
		return 0
	}

	// Weighted score based on different metrics
	viewScore := float64(s.Views) * 0.3
	engagementScore := s.CalculateEngagementRate() * 10 * 0.4
	watchTimeScore := s.AvgWatchTime * 0.2
	revenueScore := s.Revenue * 0.1

	return viewScore + engagementScore + watchTimeScore + revenueScore
}