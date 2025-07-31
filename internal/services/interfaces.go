package services

import (
	"context"
	"time"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

// VideoService defines the interface for video-related business logic
type VideoService interface {
	// Video CRUD operations
	CreateVideo(ctx context.Context, tenantID, userID string, req *models.CreateVideoRequest) (*models.Video, error)
	GetVideo(ctx context.Context, tenantID, videoID string) (*models.Video, error)
	UpdateVideo(ctx context.Context, tenantID, videoID string, req *models.UpdateVideoRequest) (*models.Video, error)
	DeleteVideo(ctx context.Context, tenantID, videoID string) error
	ListVideos(ctx context.Context, tenantID string, limit, offset int) ([]*models.Video, error)
	GetUserVideos(ctx context.Context, tenantID, userID string, limit, offset int) ([]*models.Video, error)

	// Video processing operations
	UploadVideo(ctx context.Context, tenantID, videoID string, fileData []byte, filename string) error
	UpdateVideoStatus(ctx context.Context, tenantID, videoID string, status models.VideoStatus) error
	GetVideosByStatus(ctx context.Context, tenantID string, status models.VideoStatus, limit, offset int) ([]*models.Video, error)
	SetProcessingComplete(ctx context.Context, tenantID, videoID string, duration int, resolution, thumbnailURL, s3Key, s3Bucket string) error
	SetProcessingFailed(ctx context.Context, tenantID, videoID string) error

	// Video publishing operations
	PublishVideo(ctx context.Context, tenantID, videoID string, platforms []string) error
	GetVideoPublications(ctx context.Context, tenantID, videoID string) ([]*models.PublicationJob, error)
	UpdatePublication(ctx context.Context, tenantID, publicationID string, status string) error
	CancelPublication(ctx context.Context, tenantID, publicationID string) error
}

// AIService defines the interface for AI-related business logic
type AIService interface {
	// Magic Brush operations (real-time AI generation)
	GenerateMagicBrush(ctx context.Context, tenantID string, req *MagicBrushRequest) (*MagicBrushResponse, error)

	// AI processing operations
	ProcessWithBedrock(ctx context.Context, promptKey string, input map[string]interface{}) (map[string]interface{}, error)
}

// AnalyticsService defines the interface for analytics and statistics business logic
type AnalyticsService interface {
	// Video statistics
	GetVideoStats(ctx context.Context, tenantID, videoID string) (*models.VideoStats, error)
	GetVideosStats(ctx context.Context, tenantID string, videoIDs []string) ([]*models.VideoStats, error)
	GetVideoStatsHistory(ctx context.Context, tenantID, videoID string, from, to time.Time) ([]*models.VideoStats, error)

	// Dashboard analytics
	GetDashboardStats(ctx context.Context, tenantID string) (*DashboardStats, error)
	GetPerformanceStats(ctx context.Context, tenantID string, from, to time.Time) (*PerformanceStats, error)

	// Advanced analytics
	GetROIAnalytics(ctx context.Context, tenantID string, from, to time.Time) (*ROIAnalytics, error)
	GetEngagementAnalytics(ctx context.Context, tenantID string, from, to time.Time) (*EngagementAnalytics, error)

	// Data synchronization
	SyncStats(ctx context.Context, tenantID string) error
}

// CampaignService defines the interface for AI campaign management
type CampaignService interface {
	// Campaign CRUD operations
	CreateCampaign(ctx context.Context, tenantID, userID string, req *CreateCampaignRequest) (*Campaign, error)
	GetCampaign(ctx context.Context, tenantID, campaignID string) (*Campaign, error)
	UpdateCampaign(ctx context.Context, tenantID, campaignID string, req *UpdateCampaignRequest) (*Campaign, error)
	DeleteCampaign(ctx context.Context, tenantID, campaignID string) error
	ListCampaigns(ctx context.Context, tenantID string, limit, offset int) ([]*Campaign, error)

	// Campaign execution operations
	StartCampaign(ctx context.Context, tenantID, campaignID string) error
	StopCampaign(ctx context.Context, tenantID, campaignID string) error
	PauseCampaign(ctx context.Context, tenantID, campaignID string) error
	ResumeCampaign(ctx context.Context, tenantID, campaignID string) error

	// Campaign workflow operations
	ExecuteResearchStep(ctx context.Context, tenantID, campaignID string) error
	ExecuteIdeationStep(ctx context.Context, tenantID, campaignID string) error
	ExecuteValidationStep(ctx context.Context, tenantID, campaignID string) error

	// Campaign scheduling operations
	ScheduleCampaign(ctx context.Context, tenantID, campaignID string, schedule *CampaignSchedule) error
	GetScheduledCampaigns(ctx context.Context, before time.Time, limit int) ([]*Campaign, error)
	ProcessScheduledCampaigns(ctx context.Context) error
}

// PromptService defines the interface for prompt catalog management
type PromptService interface {
	// Prompt retrieval operations
	GetPrompt(ctx context.Context, key string) (*Prompt, error)
	ListPrompts(ctx context.Context) ([]*Prompt, error)
	GetPromptsByCategory(ctx context.Context, category string) ([]*Prompt, error)

	// Prompt management operations
	CreatePrompt(ctx context.Context, req *CreatePromptRequest) (*Prompt, error)
	UpdatePrompt(ctx context.Context, key string, req *UpdatePromptRequest) (*Prompt, error)
	DeletePrompt(ctx context.Context, key string) error

	// Prompt rendering operations
	RenderPrompt(ctx context.Context, key string, data map[string]interface{}) (string, error)

	// Prompt validation operations
	ValidatePrompt(ctx context.Context, prompt *Prompt) error
	TestPrompt(ctx context.Context, key string, testData map[string]interface{}) (*PromptTestResult, error)
}

// Request/Response types for services

// MagicBrushRequest represents a request for magic brush generation
type MagicBrushRequest struct {
	VideoID   string                 `json:"video_id" validate:"required"`
	BrushType string                 `json:"brush_type" validate:"required,oneof=title description tags thumbnail"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Language  string                 `json:"language,omitempty" validate:"omitempty,len=2"`
	Tone      string                 `json:"tone,omitempty" validate:"omitempty,oneof=professional casual creative formal"`
	MaxLength int                    `json:"max_length,omitempty" validate:"omitempty,min=1,max=1000"`
}

// MagicBrushResponse represents the response from magic brush generation
type MagicBrushResponse struct {
	VideoID     string                 `json:"video_id"`
	BrushType   string                 `json:"brush_type"`
	Result      string                 `json:"result"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// CreateCampaignRequest represents a request to create a new campaign
type CreateCampaignRequest struct {
	Name      string                 `json:"name" validate:"required,max=255"`
	Goal      string                 `json:"goal" validate:"required,max=1000"`
	Context   map[string]interface{} `json:"context" validate:"required"`
	Theme     string                 `json:"theme,omitempty" validate:"max=255"`
	Platforms []string               `json:"platforms" validate:"required,min=1"`
	Language  string                 `json:"language" validate:"required,len=2"`
	Schedule  *CampaignSchedule      `json:"schedule,omitempty"`
	Budget    float64                `json:"budget,omitempty" validate:"omitempty,min=0"`
	MaxVideos int                    `json:"max_videos,omitempty" validate:"omitempty,min=1,max=100"`
}

// UpdateCampaignRequest represents a request to update an existing campaign
type UpdateCampaignRequest struct {
	Name      *string                `json:"name,omitempty" validate:"omitempty,max=255"`
	Goal      *string                `json:"goal,omitempty" validate:"omitempty,max=1000"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Theme     *string                `json:"theme,omitempty" validate:"omitempty,max=255"`
	Platforms []string               `json:"platforms,omitempty" validate:"omitempty,min=1"`
	Language  *string                `json:"language,omitempty" validate:"omitempty,len=2"`
	Schedule  *CampaignSchedule      `json:"schedule,omitempty"`
	Budget    *float64               `json:"budget,omitempty" validate:"omitempty,min=0"`
	MaxVideos *int                   `json:"max_videos,omitempty" validate:"omitempty,min=1,max=100"`
}

// Campaign represents an AI campaign
type Campaign struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	UserID      string                 `json:"user_id" db:"user_id"`
	Name        string                 `json:"name" db:"name"`
	Goal        string                 `json:"goal" db:"goal"`
	Context     map[string]interface{} `json:"context" db:"context"`
	Theme       string                 `json:"theme" db:"theme"`
	Platforms   []string               `json:"platforms" db:"platforms"`
	Language    string                 `json:"language" db:"language"`
	Status      CampaignStatus         `json:"status" db:"status"`
	Schedule    *CampaignSchedule      `json:"schedule,omitempty" db:"schedule"`
	Budget      float64                `json:"budget" db:"budget"`
	MaxVideos   int                    `json:"max_videos" db:"max_videos"`
	Progress    CampaignProgress       `json:"progress" db:"progress"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

// CampaignStatus represents the status of a campaign
type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusScheduled CampaignStatus = "scheduled"
	CampaignStatusRunning   CampaignStatus = "running"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusFailed    CampaignStatus = "failed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

// CampaignSchedule represents the scheduling configuration for a campaign
type CampaignSchedule struct {
	Type      ScheduleType `json:"type" validate:"required,oneof=once daily weekly monthly cron"`
	StartTime time.Time    `json:"start_time" validate:"required"`
	EndTime   *time.Time   `json:"end_time,omitempty"`
	CronExpr  string       `json:"cron_expr,omitempty" validate:"required_if=Type cron"`
	Timezone  string       `json:"timezone" validate:"required"`
	MaxRuns   int          `json:"max_runs,omitempty" validate:"omitempty,min=1"`
	RunCount  int          `json:"run_count" db:"run_count"`
	LastRunAt *time.Time   `json:"last_run_at,omitempty" db:"last_run_at"`
	NextRunAt *time.Time   `json:"next_run_at,omitempty" db:"next_run_at"`
}

// ScheduleType represents the type of campaign schedule
type ScheduleType string

const (
	ScheduleTypeOnce    ScheduleType = "once"
	ScheduleTypeDaily   ScheduleType = "daily"
	ScheduleTypeWeekly  ScheduleType = "weekly"
	ScheduleTypeMonthly ScheduleType = "monthly"
	ScheduleTypeCron    ScheduleType = "cron"
)

// CampaignProgress represents the progress of a campaign
type CampaignProgress struct {
	CurrentStep     CampaignStep `json:"current_step" db:"current_step"`
	ResearchDone    bool         `json:"research_done" db:"research_done"`
	IdeationDone    bool         `json:"ideation_done" db:"ideation_done"`
	ValidationDone  bool         `json:"validation_done" db:"validation_done"`
	VideosCreated   int          `json:"videos_created" db:"videos_created"`
	VideosPublished int          `json:"videos_published" db:"videos_published"`
	TotalCost       float64      `json:"total_cost" db:"total_cost"`
}

// CampaignStep represents the current step in a campaign workflow
type CampaignStep string

const (
	CampaignStepResearch   CampaignStep = "research"
	CampaignStepIdeation   CampaignStep = "ideation"
	CampaignStepValidation CampaignStep = "validation"
	CampaignStepExecution  CampaignStep = "execution"
	CampaignStepCompleted  CampaignStep = "completed"
)

// Analytics response types

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalVideos       int64    `json:"total_videos"`
	TotalViews        int64    `json:"total_views"`
	TotalEngagement   int64    `json:"total_engagement"`
	AverageROI        float64  `json:"average_roi"`
	ActiveCampaigns   int64    `json:"active_campaigns"`
	PendingBatches    int64    `json:"pending_batches"`
	MonthlyGrowth     float64  `json:"monthly_growth"`
	TopPerformingTags []string `json:"top_performing_tags"`
}

// PerformanceStats represents performance statistics
type PerformanceStats struct {
	Period           string                      `json:"period"`
	VideoMetrics     *VideoPerformanceMetrics    `json:"video_metrics"`
	PlatformMetrics  map[string]*PlatformMetrics `json:"platform_metrics"`
	EngagementTrends []*EngagementTrend          `json:"engagement_trends"`
}

// VideoPerformanceMetrics represents video performance metrics
type VideoPerformanceMetrics struct {
	TotalVideos     int64   `json:"total_videos"`
	AverageViews    float64 `json:"average_views"`
	AverageLikes    float64 `json:"average_likes"`
	AverageShares   float64 `json:"average_shares"`
	AverageComments float64 `json:"average_comments"`
	CompletionRate  float64 `json:"completion_rate"`
}

// PlatformMetrics represents platform-specific metrics
type PlatformMetrics struct {
	Platform       string  `json:"platform"`
	TotalVideos    int64   `json:"total_videos"`
	TotalViews     int64   `json:"total_views"`
	TotalLikes     int64   `json:"total_likes"`
	TotalShares    int64   `json:"total_shares"`
	TotalComments  int64   `json:"total_comments"`
	EngagementRate float64 `json:"engagement_rate"`
	ROI            float64 `json:"roi"`
}

// EngagementTrend represents engagement trend data
type EngagementTrend struct {
	Date       time.Time `json:"date"`
	Views      int64     `json:"views"`
	Likes      int64     `json:"likes"`
	Shares     int64     `json:"shares"`
	Comments   int64     `json:"comments"`
	Engagement float64   `json:"engagement"`
}

// ROIAnalytics represents ROI analytics data
type ROIAnalytics struct {
	Period         string           `json:"period"`
	TotalRevenue   float64          `json:"total_revenue"`
	TotalCost      float64          `json:"total_cost"`
	NetProfit      float64          `json:"net_profit"`
	ROI            float64          `json:"roi"`
	CostBreakdown  *CostBreakdown   `json:"cost_breakdown"`
	RevenueStreams []*RevenueStream `json:"revenue_streams"`
}

// CostBreakdown represents cost breakdown data
type CostBreakdown struct {
	ProductionCost float64 `json:"production_cost"`
	AICost         float64 `json:"ai_cost"`
	PlatformCost   float64 `json:"platform_cost"`
	MarketingCost  float64 `json:"marketing_cost"`
	OtherCost      float64 `json:"other_cost"`
}

// RevenueStream represents a revenue stream
type RevenueStream struct {
	Source  string  `json:"source"`
	Amount  float64 `json:"amount"`
	Percent float64 `json:"percent"`
}

// EngagementAnalytics represents engagement analytics data
type EngagementAnalytics struct {
	Period               string                         `json:"period"`
	TotalEngagement      int64                          `json:"total_engagement"`
	AverageEngagement    float64                        `json:"average_engagement"`
	EngagementRate       float64                        `json:"engagement_rate"`
	TopContent           []*ContentEngagement           `json:"top_content"`
	EngagementByPlatform map[string]*PlatformEngagement `json:"engagement_by_platform"`
	EngagementTrends     []*EngagementTrend             `json:"engagement_trends"`
}

// ContentEngagement represents content engagement data
type ContentEngagement struct {
	VideoID        string  `json:"video_id"`
	Title          string  `json:"title"`
	Views          int64   `json:"views"`
	Likes          int64   `json:"likes"`
	Shares         int64   `json:"shares"`
	Comments       int64   `json:"comments"`
	EngagementRate float64 `json:"engagement_rate"`
}

// PlatformEngagement represents platform engagement data
type PlatformEngagement struct {
	Platform       string  `json:"platform"`
	TotalViews     int64   `json:"total_views"`
	TotalLikes     int64   `json:"total_likes"`
	TotalShares    int64   `json:"total_shares"`
	TotalComments  int64   `json:"total_comments"`
	EngagementRate float64 `json:"engagement_rate"`
}

// Prompt catalog types

// Prompt represents a prompt in the catalog
type Prompt struct {
	Key         string                 `json:"key" yaml:"key"`
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description" yaml:"description"`
	Category    string                 `json:"category" yaml:"category"`
	Template    string                 `json:"template" yaml:"template"`
	Variables   []PromptVariable       `json:"variables" yaml:"variables"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Version     string                 `json:"version" yaml:"version"`
	CreatedAt   time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" yaml:"updated_at"`
}

// PromptVariable represents a variable in a prompt template
type PromptVariable struct {
	Name        string      `json:"name" yaml:"name"`
	Type        string      `json:"type" yaml:"type"`
	Description string      `json:"description" yaml:"description"`
	Required    bool        `json:"required" yaml:"required"`
	Default     interface{} `json:"default,omitempty" yaml:"default,omitempty"`
	Validation  string      `json:"validation,omitempty" yaml:"validation,omitempty"`
}

// CreatePromptRequest represents a request to create a new prompt
type CreatePromptRequest struct {
	Key         string                 `json:"key" validate:"required,max=100"`
	Name        string                 `json:"name" validate:"required,max=255"`
	Description string                 `json:"description" validate:"required,max=1000"`
	Category    string                 `json:"category" validate:"required,max=100"`
	Template    string                 `json:"template" validate:"required"`
	Variables   []PromptVariable       `json:"variables,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdatePromptRequest represents a request to update an existing prompt
type UpdatePromptRequest struct {
	Name        *string                `json:"name,omitempty" validate:"omitempty,max=255"`
	Description *string                `json:"description,omitempty" validate:"omitempty,max=1000"`
	Category    *string                `json:"category,omitempty" validate:"omitempty,max=100"`
	Template    *string                `json:"template,omitempty"`
	Variables   []PromptVariable       `json:"variables,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PromptTestResult represents the result of testing a prompt
type PromptTestResult struct {
	Success    bool                   `json:"success"`
	Result     string                 `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration"`
	TokensUsed int                    `json:"tokens_used,omitempty"`
	Cost       float64                `json:"cost,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	TestedAt   time.Time              `json:"tested_at"`
}
