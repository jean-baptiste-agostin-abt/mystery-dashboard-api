package logger

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
}

// New creates a new logger instance with the specified log level and environment
func New(logLevel, environment string) *Logger {
	var config zap.Config

	// Configure logger based on environment
	if environment == "production" {
		config = zap.NewProductionConfig()
		config.DisableStacktrace = true
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		level = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)

	// Add caller information
	config.DisableCaller = false
	config.DisableStacktrace = environment == "production"

	// Build logger
	zapLogger, err := config.Build(
		zap.AddCallerSkip(1), // Skip one level to show the actual caller
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	return &Logger{Logger: zapLogger}
}

// WithContext adds trace information to the logger if available
func (l *Logger) WithContext(ctx context.Context) *Logger {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return l
	}

	spanContext := span.SpanContext()
	return &Logger{
		Logger: l.Logger.With(
			zap.String("trace_id", spanContext.TraceID().String()),
			zap.String("span_id", spanContext.SpanID().String()),
		),
	}
}

// WithTenant adds tenant information to the logger
func (l *Logger) WithTenant(tenantID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("tenant_id", tenantID)),
	}
}

// WithUser adds user information to the logger
func (l *Logger) WithUser(userID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("user_id", userID)),
	}
}

// WithRequestID adds request ID to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("request_id", requestID)),
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return &Logger{
		Logger: l.Logger.With(zapFields...),
	}
}

// Debug logs a debug message with optional fields
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.Logger.Debug(msg, l.parseFields(fields...)...)
}

// Info logs an info message with optional fields
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.Logger.Info(msg, l.parseFields(fields...)...)
}

// Warn logs a warning message with optional fields
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.Logger.Warn(msg, l.parseFields(fields...)...)
}

// Error logs an error message with optional fields
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.Logger.Error(msg, l.parseFields(fields...)...)
}

// Fatal logs a fatal message with optional fields and exits
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.Logger.Fatal(msg, l.parseFields(fields...)...)
}

// parseFields converts key-value pairs to zap.Field
func (l *Logger) parseFields(fields ...interface{}) []zap.Field {
	if len(fields)%2 != 0 {
		// If odd number of fields, add the last one as a generic field
		fields = append(fields, "MISSING_VALUE")
	}

	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			key = fmt.Sprintf("field_%d", i/2)
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}

	return zapFields
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// HTTPMiddleware creates a middleware function for HTTP request logging
func (l *Logger) HTTPMiddleware() func(ctx context.Context, method, path string, statusCode int, duration float64, requestID string) {
	return func(ctx context.Context, method, path string, statusCode int, duration float64, requestID string) {
		logger := l.WithContext(ctx).WithRequestID(requestID)
		
		fields := []interface{}{
			"method", method,
			"path", path,
			"status_code", statusCode,
			"duration_ms", duration,
		}

		if statusCode >= 500 {
			logger.Error("HTTP request completed with server error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP request completed with client error", fields...)
		} else {
			logger.Info("HTTP request completed", fields...)
		}
	}
}