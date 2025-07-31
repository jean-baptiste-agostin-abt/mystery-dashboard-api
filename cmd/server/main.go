package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/mysteryfactory/internal/config"
	"github.com/yourorg/mysteryfactory/internal/router"
	"github.com/yourorg/mysteryfactory/pkg/db"
	"github.com/yourorg/mysteryfactory/pkg/logger"
	"github.com/yourorg/mysteryfactory/pkg/metrics"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// @title Mystery Factory API
// @version 1.0
// @description A multi-tenant video management and publishing platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.mysteryfactory.io/support
// @contact.email support@mysteryfactory.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.LogLevel, cfg.Environment)
	defer logger.Sync()

	// Initialize OpenTelemetry
	tp, err := initTracer(cfg.ServiceName, cfg.JaegerEndpoint)
	if err != nil {
		logger.Fatal("Failed to initialize tracer", "error", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer provider", "error", err)
		}
	}()

	// Initialize Prometheus metrics
	m := metrics.New()

	// Initialize database
	database, err := db.New(cfg.DatabaseDSN)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(cfg.DatabaseDSN); err != nil {
		logger.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize router
	r := router.New(cfg, logger, database, m)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", "port", cfg.Port, "environment", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer(serviceName, jaegerEndpoint string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	return tp, nil
}
