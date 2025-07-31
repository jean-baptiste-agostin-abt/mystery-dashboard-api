package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// AI processing metrics
	AIRequestsTotal    *prometheus.CounterVec
	AIRequestDuration  *prometheus.HistogramVec
	AITokensUsed       *prometheus.CounterVec
	AIRequestsInFlight prometheus.Gauge

	// Database metrics
	DBConnectionsActive prometheus.Gauge
	DBConnectionsIdle   prometheus.Gauge
	DBQueriesTotal      *prometheus.CounterVec
	DBQueryDuration     *prometheus.HistogramVec

	// Business metrics
	VideosTotal         *prometheus.CounterVec
	VideoProcessingTime *prometheus.HistogramVec
	CampaignsTotal      *prometheus.CounterVec
	CampaignSuccess     *prometheus.CounterVec
	MagicBrushRequests  *prometheus.CounterVec

	// System metrics
	ErrorsTotal *prometheus.CounterVec
	PanicTotal  prometheus.Counter
}

// New creates and registers all Prometheus metrics
func New() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code", "tenant_id"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code", "tenant_id"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		// AI processing metrics
		AIRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_requests_total",
				Help: "Total number of AI requests",
			},
			[]string{"model", "prompt_key", "brush_type", "status", "tenant_id"},
		),
		AIRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ai_request_duration_seconds",
				Help:    "Duration of AI requests in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0},
			},
			[]string{"model", "prompt_key", "brush_type", "tenant_id"},
		),
		AITokensUsed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_tokens_used_total",
				Help: "Total number of AI tokens used",
			},
			[]string{"model", "prompt_key", "type", "tenant_id"},
		),
		AIRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ai_requests_in_flight",
				Help: "Number of AI requests currently being processed",
			},
		),

		// Database metrics
		DBConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_idle",
				Help: "Number of idle database connections",
			},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status", "tenant_id"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
			},
			[]string{"operation", "table", "tenant_id"},
		),

		// Business metrics
		VideosTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "videos_total",
				Help: "Total number of videos",
			},
			[]string{"status", "tenant_id"},
		),
		VideoProcessingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "video_processing_duration_seconds",
				Help:    "Duration of video processing in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
			},
			[]string{"status", "tenant_id"},
		),
		CampaignsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "campaigns_total",
				Help: "Total number of campaigns",
			},
			[]string{"status", "tenant_id"},
		),
		CampaignSuccess: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "campaign_success_total",
				Help: "Total number of successful campaign operations",
			},
			[]string{"operation", "tenant_id"},
		),
		MagicBrushRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "magic_brush_requests_total",
				Help: "Total number of magic brush requests",
			},
			[]string{"brush_type", "status", "tenant_id"},
		),

		// System metrics
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Total number of errors",
			},
			[]string{"type", "component", "tenant_id"},
		),
		PanicTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "panics_total",
				Help: "Total number of panics",
			},
		),
	}
}

// HTTPMiddleware returns a Gin middleware for HTTP metrics collection
func (m *Metrics) HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "unknown"
		}

		// Increment in-flight requests
		m.HTTPRequestsInFlight.Inc()
		defer m.HTTPRequestsInFlight.Dec()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		labels := prometheus.Labels{
			"method":      c.Request.Method,
			"endpoint":    c.FullPath(),
			"status_code": statusCode,
			"tenant_id":   tenantID,
		}

		m.HTTPRequestsTotal.With(labels).Inc()
		m.HTTPRequestDuration.With(labels).Observe(duration)
	}
}

// RecordAIRequest records metrics for AI requests
func (m *Metrics) RecordAIRequest(model, promptKey, brushType, status, tenantID string, duration time.Duration, tokensUsed int) {
	labels := prometheus.Labels{
		"model":      model,
		"prompt_key": promptKey,
		"brush_type": brushType,
		"status":     status,
		"tenant_id":  tenantID,
	}

	m.AIRequestsTotal.With(labels).Inc()

	durationLabels := prometheus.Labels{
		"model":      model,
		"prompt_key": promptKey,
		"brush_type": brushType,
		"tenant_id":  tenantID,
	}
	m.AIRequestDuration.With(durationLabels).Observe(duration.Seconds())

	if tokensUsed > 0 {
		tokenLabels := prometheus.Labels{
			"model":      model,
			"prompt_key": promptKey,
			"type":       "total",
			"tenant_id":  tenantID,
		}
		m.AITokensUsed.With(tokenLabels).Add(float64(tokensUsed))
	}
}

// RecordDBQuery records metrics for database queries
func (m *Metrics) RecordDBQuery(operation, table, status, tenantID string, duration time.Duration) {
	queryLabels := prometheus.Labels{
		"operation": operation,
		"table":     table,
		"status":    status,
		"tenant_id": tenantID,
	}
	m.DBQueriesTotal.With(queryLabels).Inc()

	durationLabels := prometheus.Labels{
		"operation": operation,
		"table":     table,
		"tenant_id": tenantID,
	}
	m.DBQueryDuration.With(durationLabels).Observe(duration.Seconds())
}

// RecordVideo records metrics for video operations
func (m *Metrics) RecordVideo(status, tenantID string) {
	labels := prometheus.Labels{
		"status":    status,
		"tenant_id": tenantID,
	}
	m.VideosTotal.With(labels).Inc()
}

// RecordVideoProcessing records metrics for video processing duration
func (m *Metrics) RecordVideoProcessing(status, tenantID string, duration time.Duration) {
	labels := prometheus.Labels{
		"status":    status,
		"tenant_id": tenantID,
	}
	m.VideoProcessingTime.With(labels).Observe(duration.Seconds())
}

// RecordCampaign records metrics for campaign operations
func (m *Metrics) RecordCampaign(status, tenantID string) {
	labels := prometheus.Labels{
		"status":    status,
		"tenant_id": tenantID,
	}
	m.CampaignsTotal.With(labels).Inc()
}

// RecordCampaignSuccess records metrics for successful campaign operations
func (m *Metrics) RecordCampaignSuccess(operation, tenantID string) {
	labels := prometheus.Labels{
		"operation": operation,
		"tenant_id": tenantID,
	}
	m.CampaignSuccess.With(labels).Inc()
}

// RecordMagicBrush records metrics for magic brush requests
func (m *Metrics) RecordMagicBrush(brushType, status, tenantID string) {
	labels := prometheus.Labels{
		"brush_type": brushType,
		"status":     status,
		"tenant_id":  tenantID,
	}
	m.MagicBrushRequests.With(labels).Inc()
}

// RecordError records metrics for errors
func (m *Metrics) RecordError(errorType, component, tenantID string) {
	labels := prometheus.Labels{
		"type":      errorType,
		"component": component,
		"tenant_id": tenantID,
	}
	m.ErrorsTotal.With(labels).Inc()
}

// RecordPanic records metrics for panics
func (m *Metrics) RecordPanic() {
	m.PanicTotal.Inc()
}

// UpdateDBConnections updates database connection metrics
func (m *Metrics) UpdateDBConnections(active, idle int) {
	m.DBConnectionsActive.Set(float64(active))
	m.DBConnectionsIdle.Set(float64(idle))
}

// IncrementAIInFlight increments AI requests in flight
func (m *Metrics) IncrementAIInFlight() {
	m.AIRequestsInFlight.Inc()
}

// DecrementAIInFlight decrements AI requests in flight
func (m *Metrics) DecrementAIInFlight() {
	m.AIRequestsInFlight.Dec()
}
