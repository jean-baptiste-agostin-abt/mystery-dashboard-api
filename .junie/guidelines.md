# Mystery Factory API - AI Agent Guidelines

## Architectural Objectives

### 1. Multi-Tenant Architecture
**Objective**: Provide secure, scalable tenant isolation with efficient resource sharing.

**Implementation Standards**:
- All database queries must include tenant context filtering
- JWT tokens must contain tenant information for request validation
- API endpoints must validate tenant access permissions
- Database schema must support tenant-specific data isolation
- Shared resources (caching, queuing) must be tenant-aware

**Key Principles**:
- Tenant data must never leak between tenants
- Performance isolation to prevent tenant interference
- Scalable tenant onboarding without architectural changes
- Cost-effective resource sharing while maintaining security

### 2. Event-Driven Architecture
**Objective**: Enable loose coupling, scalability, and real-time processing through event-driven patterns.

**Implementation Standards**:
- Use domain events for business logic decoupling
- Implement event sourcing for audit trails and state reconstruction
- Asynchronous processing for non-critical operations
- Event-driven communication between bounded contexts
- Reliable event delivery with retry mechanisms

**Event Categories**:
- **Domain Events**: Business logic changes (VideoUploaded, PublicationCompleted)
- **Integration Events**: External system interactions (PlatformWebhookReceived)
- **System Events**: Infrastructure concerns (DatabaseConnectionLost)

### 3. Observability First
**Objective**: Comprehensive monitoring, tracing, and debugging capabilities from day one.

**Implementation Standards**:
- OpenTelemetry integration for distributed tracing
- Structured logging with correlation IDs
- Metrics collection for business and technical KPIs
- Health checks for all dependencies
- Error tracking with context preservation

**Observability Pillars**:
- **Logs**: Structured, searchable, with proper log levels
- **Metrics**: Business metrics, performance metrics, infrastructure metrics
- **Traces**: End-to-end request tracing across services
- **Alerts**: Proactive monitoring with actionable alerts

## Code Standards and Best Practices

### 1. Go Language Standards

#### Naming Conventions
```go
// Package names: lowercase, single word
package handlers

// Types: PascalCase
type VideoHandler struct {}

// Functions/Methods: PascalCase for exported, camelCase for unexported
func (h *VideoHandler) CreateVideo() {}
func (h *VideoHandler) validateInput() {}

// Variables: camelCase
var userID string
var maxRetryCount int

// Constants: PascalCase or SCREAMING_SNAKE_CASE for package-level
const MaxUploadSize = 1024 * 1024 * 100
const DEFAULT_TIMEOUT = 30
```

#### Error Handling
```go
// Always handle errors explicitly
result, err := someOperation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use custom error types for domain errors
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Wrap errors with context
if err := validateUser(user); err != nil {
    return fmt.Errorf("user validation failed for ID %s: %w", user.ID, err)
}
```

#### Struct Design
```go
// Use composition over inheritance
type BaseEntity struct {
    ID        string    `json:"id" db:"id"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Video struct {
    BaseEntity
    TenantID    string `json:"tenant_id" db:"tenant_id"`
    Title       string `json:"title" db:"title"`
    Description string `json:"description" db:"description"`
}

// Use builder pattern for complex objects
type VideoBuilder struct {
    video *Video
}

func NewVideoBuilder() *VideoBuilder {
    return &VideoBuilder{video: &Video{}}
}

func (b *VideoBuilder) WithTitle(title string) *VideoBuilder {
    b.video.Title = title
    return b
}

func (b *VideoBuilder) Build() *Video {
    return b.video
}
```

### 2. API Design Standards

#### RESTful Resource Design
```
GET    /api/v1/videos           # List videos
POST   /api/v1/videos           # Create video
GET    /api/v1/videos/{id}      # Get video
PUT    /api/v1/videos/{id}      # Update video (full)
PATCH  /api/v1/videos/{id}      # Update video (partial)
DELETE /api/v1/videos/{id}      # Delete video

# Nested resources
GET    /api/v1/videos/{id}/publications
POST   /api/v1/videos/{id}/publications
GET    /api/v1/videos/{id}/publications/{pub_id}
```

#### Response Format Standards
```go
// Success Response
type SuccessResponse struct {
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Error Response
type ErrorResponse struct {
    Error   string            `json:"error"`
    Message string            `json:"message"`
    Code    int               `json:"code,omitempty"`
    Details map[string]string `json:"details,omitempty"`
}

// Paginated Response
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Total      int64       `json:"total"`
    Page       int         `json:"page"`
    Limit      int         `json:"limit"`
    TotalPages int         `json:"total_pages"`
}
```

#### HTTP Status Code Usage
- `200 OK`: Successful GET, PUT, PATCH
- `201 Created`: Successful POST
- `204 No Content`: Successful DELETE
- `400 Bad Request`: Client error, validation failure
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict
- `422 Unprocessable Entity`: Validation error with details
- `500 Internal Server Error`: Server error

### 3. Database Design Standards

#### Schema Design Principles
```sql
-- Use consistent naming conventions
CREATE TABLE tenants (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    settings JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Always include tenant_id for multi-tenancy
CREATE TABLE videos (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status ENUM('uploading', 'processing', 'ready', 'failed') DEFAULT 'uploading',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

#### Migration Standards
```go
// Migration file naming: YYYYMMDDHHMMSS_description.sql
// 20240131120000_create_videos_table.up.sql

-- +migrate Up
CREATE TABLE videos (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    -- ... other fields
);

-- +migrate Down
DROP TABLE videos;
```

### 4. Testing Standards

#### Unit Test Structure
```go
func TestVideoHandler_CreateVideo(t *testing.T) {
    tests := []struct {
        name           string
        requestBody    interface{}
        expectedStatus int
        expectedError  string
        setupMocks     func(*mocks.MockVideoService)
    }{
        {
            name: "successful video creation",
            requestBody: models.CreateVideoRequest{
                Title:    "Test Video",
                FileName: "test.mp4",
                FileSize: 1024000,
                Format:   "mp4",
            },
            expectedStatus: http.StatusCreated,
            setupMocks: func(m *mocks.MockVideoService) {
                m.EXPECT().CreateVideo(gomock.Any(), gomock.Any(), gomock.Any()).
                    Return(&models.Video{ID: "video-123"}, nil)
            },
        },
        {
            name: "validation error - missing title",
            requestBody: models.CreateVideoRequest{
                FileName: "test.mp4",
                FileSize: 1024000,
                Format:   "mp4",
            },
            expectedStatus: http.StatusBadRequest,
            expectedError:  "title is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Integration Test Standards
```go
func TestVideoIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Setup test server
    server := setupTestServer(t, db)
    defer server.Close()
    
    // Test scenarios
    t.Run("complete video workflow", func(t *testing.T) {
        // Create video -> Upload file -> Publish -> Verify stats
    })
}
```

## Security Standards

### 1. Authentication & Authorization

#### JWT Token Structure
```go
type JWTClaims struct {
    UserID   string   `json:"user_id"`
    TenantID string   `json:"tenant_id"`
    Email    string   `json:"email"`
    Role     string   `json:"role"`
    Permissions []string `json:"permissions"`
    jwt.RegisteredClaims
}
```

#### Permission-Based Access Control
```go
// Define permissions as constants
const (
    PermissionReadVideos   = "videos:read"
    PermissionWriteVideos  = "videos:write"
    PermissionDeleteVideos = "videos:delete"
    PermissionPublishVideos = "videos:publish"
)

// Role-based permission mapping
var RolePermissions = map[string][]string{
    "admin":     {"*"}, // All permissions
    "editor":    {PermissionReadVideos, PermissionWriteVideos, PermissionPublishVideos},
    "viewer":    {PermissionReadVideos},
    "publisher": {PermissionReadVideos, PermissionPublishVideos},
}
```

### 2. Input Validation & Sanitization

#### Request Validation
```go
type CreateVideoRequest struct {
    Title       string   `json:"title" validate:"required,max=255"`
    Description string   `json:"description" validate:"max=1000"`
    FileName    string   `json:"file_name" validate:"required"`
    FileSize    int64    `json:"file_size" validate:"required,min=1,max=104857600"` // 100MB
    Format      string   `json:"format" validate:"required,oneof=mp4 avi mov wmv"`
    Tags        []string `json:"tags,omitempty" validate:"max=10,dive,max=50"`
}

func (r *CreateVideoRequest) Validate() error {
    validate := validator.New()
    return validate.Struct(r)
}
```

#### SQL Injection Prevention
```go
// Always use parameterized queries
func (r *VideoRepository) GetByTenantAndUser(tenantID, userID string) ([]*Video, error) {
    query := `
        SELECT id, tenant_id, user_id, title, description, status, created_at
        FROM videos 
        WHERE tenant_id = ? AND user_id = ? AND deleted_at IS NULL
        ORDER BY created_at DESC
    `
    
    rows, err := r.db.Query(query, tenantID, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to query videos: %w", err)
    }
    defer rows.Close()
    
    // Process results...
}
```

## Performance Standards

### 1. Database Performance

#### Query Optimization
```go
// Use appropriate indexes
CREATE INDEX idx_videos_tenant_status ON videos(tenant_id, status);
CREATE INDEX idx_videos_created_at ON videos(created_at);

// Implement pagination
func (r *VideoRepository) List(tenantID string, limit, offset int) ([]*Video, error) {
    query := `
        SELECT id, title, status, created_at
        FROM videos 
        WHERE tenant_id = ? AND deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `
    // Implementation...
}
```

#### Connection Pooling
```go
func NewDB(dsn string) (*DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return &DB{DB: db}, nil
}
```

### 2. Caching Strategy

#### Redis Caching
```go
type CacheService struct {
    redis *redis.Client
}

func (c *CacheService) GetVideoStats(videoID string) (*VideoStats, error) {
    key := fmt.Sprintf("video:stats:%s", videoID)
    
    // Try cache first
    cached, err := c.redis.Get(context.Background(), key).Result()
    if err == nil {
        var stats VideoStats
        if err := json.Unmarshal([]byte(cached), &stats); err == nil {
            return &stats, nil
        }
    }
    
    // Cache miss - fetch from database and cache
    stats, err := c.fetchVideoStatsFromDB(videoID)
    if err != nil {
        return nil, err
    }
    
    // Cache for 5 minutes
    statsJSON, _ := json.Marshal(stats)
    c.redis.Set(context.Background(), key, statsJSON, 5*time.Minute)
    
    return stats, nil
}
```

## Deployment and Operations

### 1. Configuration Management

#### Environment-based Configuration
```go
type Config struct {
    // Server
    Port         int    `mapstructure:"PORT" validate:"required,min=1,max=65535"`
    Environment  string `mapstructure:"ENVIRONMENT" validate:"required,oneof=development staging production"`
    
    // Database
    DatabaseDSN  string `mapstructure:"DATABASE_DSN" validate:"required"`
    
    // Security
    JWTSecret    string `mapstructure:"JWT_SECRET" validate:"required,min=32"`
    
    // External Services
    AWSRegion    string `mapstructure:"AWS_REGION" validate:"required"`
    S3Bucket     string `mapstructure:"S3_BUCKET" validate:"required"`
}

func Load() (*Config, error) {
    viper.AutomaticEnv()
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    validate := validator.New()
    if err := validate.Struct(&config); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    return &config, nil
}
```

### 2. Health Checks and Monitoring

#### Health Check Implementation
```go
type HealthChecker struct {
    db    *sql.DB
    redis *redis.Client
}

func (h *HealthChecker) Check(ctx context.Context) map[string]interface{} {
    result := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now().UTC(),
        "checks":    make(map[string]interface{}),
    }
    
    // Database health
    if err := h.db.PingContext(ctx); err != nil {
        result["checks"]["database"] = map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
        result["status"] = "unhealthy"
    } else {
        result["checks"]["database"] = map[string]interface{}{
            "status": "healthy",
        }
    }
    
    // Redis health
    if err := h.redis.Ping(ctx).Err(); err != nil {
        result["checks"]["redis"] = map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
        result["status"] = "unhealthy"
    } else {
        result["checks"]["redis"] = map[string]interface{}{
            "status": "healthy",
        }
    }
    
    return result
}
```

## AI Agent Specific Guidelines

### 1. Code Generation Standards

When generating code, always:
- Include comprehensive error handling
- Add appropriate logging with context
- Include input validation
- Follow the established patterns in the codebase
- Add unit tests for new functionality
- Update documentation as needed

### 2. Refactoring Guidelines

When refactoring existing code:
- Maintain backward compatibility for public APIs
- Update tests to reflect changes
- Preserve existing functionality
- Improve performance and maintainability
- Follow the boy scout rule: leave code better than you found it

### 3. Documentation Standards

When updating documentation:
- Keep it current with code changes
- Use clear, concise language
- Include code examples where appropriate
- Update API documentation for endpoint changes
- Maintain consistency with existing documentation style

---

These guidelines serve as the foundation for maintaining code quality, security, and performance standards throughout the Mystery Factory API project. All code contributions should adhere to these standards to ensure consistency and maintainability.