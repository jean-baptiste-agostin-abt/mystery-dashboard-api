package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jibe0123/mysteryfactory/internal/config"
	"github.com/jibe0123/mysteryfactory/internal/handlers"
	"github.com/jibe0123/mysteryfactory/internal/middleware"
	"github.com/jibe0123/mysteryfactory/internal/services"
	"github.com/jibe0123/mysteryfactory/pkg/aws"
	"github.com/jibe0123/mysteryfactory/pkg/db"
	"github.com/jibe0123/mysteryfactory/pkg/logger"
	"github.com/jibe0123/mysteryfactory/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// New creates a new Gin router with all routes and middleware configured
func New(cfg *config.Config, logger *logger.Logger, db *db.DB, metrics *metrics.Metrics) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(otelgin.Middleware(cfg.ServiceName))
	r.Use(middleware.RateLimiter())
	r.Use(metrics.HTTPMiddleware())

	// Health check endpoint (no auth required)
	r.GET("/health", handlers.HealthCheck(db))
	r.GET("/ready", handlers.ReadinessCheck(db))

	// Metrics endpoint for Prometheus
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation (only in non-production)
	if cfg.Environment != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Initialize services
	promptService, err := services.NewPromptService("prompts/catalog.yaml", logger)
	if err != nil {
		logger.Error("Failed to initialize prompt service", "error", err)
		panic(err)
	}

	bedrockClient, err := aws.NewBedrockClient(nil, logger)
	if err != nil {
		logger.Error("Failed to initialize Bedrock client", "error", err)
		panic(err)
	}

	aiService := services.NewAIService(promptService, bedrockClient, logger, metrics)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg, logger, db)
	videoHandler := handlers.NewVideoHandler(cfg, logger, db)
	platformHandler := handlers.NewPlatformHandler(cfg, logger, db)
	statsHandler := handlers.NewStatsHandler(cfg, logger, db)
	aiHandler := handlers.NewAIHandler(aiService, logger)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.JWTAuth(cfg.JWTSecret), authHandler.Logout)
			auth.GET("/me", middleware.JWTAuth(cfg.JWTSecret), authHandler.GetProfile)
			auth.PUT("/me", middleware.JWTAuth(cfg.JWTSecret), authHandler.UpdateProfile)
			auth.POST("/change-password", middleware.JWTAuth(cfg.JWTSecret), authHandler.ChangePassword)
		}

		// Protected routes (require authentication)
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(cfg.JWTSecret))
		protected.Use(middleware.TenantResolver())
		{
			// Video management routes
			videos := protected.Group("/videos")
			{
				videos.GET("", videoHandler.ListVideos)
				videos.POST("", videoHandler.CreateVideo)
				videos.GET("/:id", videoHandler.GetVideo)
				videos.PUT("/:id", videoHandler.UpdateVideo)
				videos.DELETE("/:id", videoHandler.DeleteVideo)
				videos.POST("/:id/upload", videoHandler.UploadVideo)
				videos.GET("/:id/stats", statsHandler.GetVideoStats)

				// Publication routes
				videos.POST("/:id/publish", videoHandler.PublishVideo)
				videos.GET("/:id/publications", videoHandler.GetVideoPublications)
				videos.PUT("/:id/publications/:pub_id", videoHandler.UpdatePublication)
				videos.DELETE("/:id/publications/:pub_id", videoHandler.CancelPublication)
			}

			// Platform webhook routes (special auth handling)
			platforms := protected.Group("/platforms")
			{
				platforms.POST("/webhook/:platform", platformHandler.HandleWebhook)
				platforms.GET("/webhook/:platform/verify", platformHandler.VerifyWebhook)
				platforms.GET("/:platform/auth", platformHandler.InitiatePlatformAuth)
				platforms.POST("/:platform/auth/callback", platformHandler.HandleAuthCallback)
				platforms.DELETE("/:platform/auth", platformHandler.RevokePlatformAuth)
			}

			// Statistics and analytics routes
			stats := protected.Group("/stats")
			{
				stats.GET("/videos", statsHandler.GetVideosStats)
				stats.GET("/videos/:id", statsHandler.GetVideoStats)
				stats.GET("/videos/:id/history", statsHandler.GetVideoStatsHistory)
				stats.GET("/dashboard", statsHandler.GetDashboardStats)
				stats.GET("/performance", statsHandler.GetPerformanceStats)
				stats.POST("/sync", statsHandler.SyncStats)

				// Enhanced analytics - ROI and engagement tracking
				stats.GET("/roi", statsHandler.GetROIAnalytics)
				stats.GET("/engagement", statsHandler.GetEngagementAnalytics)
			}

			// AI processing routes
			ai := protected.Group("/ai")
			{
				// Magic Brush - real-time AI content generation
				ai.POST("/magic-brush", aiHandler.GenerateMagicBrush)

				// Prompt management
				ai.GET("/prompts", aiHandler.GetPrompts)
				ai.POST("/test-prompt", aiHandler.TestPrompt)
			}

			// User management routes (admin only)
			users := protected.Group("/users")
			users.Use(middleware.RequireRole("admin"))
			{
				users.GET("", authHandler.ListUsers)
				users.POST("", authHandler.CreateUser)
				users.GET("/:id", authHandler.GetUser)
				users.PUT("/:id", authHandler.UpdateUser)
				users.DELETE("/:id", authHandler.DeleteUser)
			}

			// Tenant management routes (admin only)
			tenants := protected.Group("/tenants")
			tenants.Use(middleware.RequireRole("admin"))
			{
				tenants.GET("", authHandler.ListTenants)
				tenants.POST("", authHandler.CreateTenant)
				tenants.GET("/:id", authHandler.GetTenant)
				tenants.PUT("/:id", authHandler.UpdateTenant)
				tenants.DELETE("/:id", authHandler.DeleteTenant)
			}
		}
	}

	// Webhook routes (special handling, no standard auth)
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/:platform", middleware.WebhookAuth(), platformHandler.HandleWebhook)
	}

	// 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "The requested resource was not found",
			"path":    c.Request.URL.Path,
		})
	})

	// 405 handler
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error":   "Method Not Allowed",
			"message": "The requested method is not allowed for this resource",
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
		})
	})

	return r
}

// SetupRoutes is an alternative function for setting up routes with more granular control
func SetupRoutes(r *gin.Engine, cfg *config.Config, logger *logger.Logger, db *db.DB) {
	// This function can be used if you need more control over route setup
	// Currently, the New function handles everything, but this provides flexibility
}

// RouteInfo represents information about a route
type RouteInfo struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Handler     string   `json:"handler"`
	Middleware  []string `json:"middleware"`
	Description string   `json:"description"`
}

// GetRouteInfo returns information about all registered routes
func GetRouteInfo(r *gin.Engine) []RouteInfo {
	var routes []RouteInfo

	for _, route := range r.Routes() {
		routeInfo := RouteInfo{
			Method:  route.Method,
			Path:    route.Path,
			Handler: route.Handler,
		}
		routes = append(routes, routeInfo)
	}

	return routes
}

// RegisterCustomRoutes allows for registering additional custom routes
func RegisterCustomRoutes(r *gin.Engine, customRoutes func(*gin.Engine)) {
	if customRoutes != nil {
		customRoutes(r)
	}
}
