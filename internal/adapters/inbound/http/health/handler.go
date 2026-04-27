package health

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	appConfig "github.com/bbridges_11/document-registry/internal/platform/config"
	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// HealthChecker is a function that checks the health of a dependency
type HealthChecker func(context.Context) error

// Handler handles health and readiness endpoints for ECS
type Handler struct {
	runner    *dbPostgres.Runner
	s3Client  *s3.Client
	snsClient *sns.Client
	awsConfig appConfig.AWSConfig
	logger    *zap.Logger
}

func NewHandler(
	runner *dbPostgres.Runner,
	s3Client *s3.Client,
	snsClient *sns.Client,
	awsConfig appConfig.AWSConfig,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		runner:    runner,
		s3Client:  s3Client,
		snsClient: snsClient,
		awsConfig: awsConfig,
		logger:    logger,
	}
}

// RegisterRoutes registers health-related routes for ECS
func (h *Handler) RegisterRoutes(e *echo.Echo, middleWareFuncs ...echo.MiddlewareFunc) {
	e.GET("/health", h.Health)
	e.GET("/health/live", h.Liveness)
	e.GET("/health/ready", h.Readiness)
	e.GET("/health/startup", h.Startup)
}

// Health returns ALB target group health check
// Checks critical dependencies: Postgres + S3
// Used by ALB to determine if container can receive traffic
func (h *Handler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	results := make(map[string]string)
	allHealthy := true

	// Check critical dependencies for ALB health
	// Postgres - required for all operations
	if err := h.checkPostgres(ctx); err != nil {
		results["postgres"] = "unhealthy: " + err.Error()
		allHealthy = false
		h.logger.Warn("postgres health check failed", zap.Error(err))
	} else {
		results["postgres"] = "healthy"
	}

	// S3 - required for content storage
	if err := h.checkS3(ctx); err != nil {
		results["s3"] = "unhealthy: " + err.Error()
		allHealthy = false
		h.logger.Warn("s3 health check failed", zap.Error(err))
	} else {
		results["s3"] = "healthy"
	}

	response := map[string]any{
		"status":    "healthy",
		"checks":    results,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if !allHealthy {
		response["status"] = "unhealthy"
		return c.JSON(503, response)
	}

	return c.JSON(200, response)
}

// Liveness returns ECS container liveness probe
// Always returns 200 if process is running
// Used by Docker HEALTHCHECK
func (h *Handler) Liveness(c echo.Context) error {
	return c.JSON(200, map[string]any{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Readiness returns comprehensive readiness check
// Checks all dependencies: Postgres + S3 + SNS (if enabled)
// Used by monitoring tools and dashboards
func (h *Handler) Readiness(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	results := make(map[string]string)
	allHealthy := true

	// Check PostgreSQL
	if err := h.checkPostgres(ctx); err != nil {
		results["postgres"] = "unhealthy: " + err.Error()
		allHealthy = false
		h.logger.Warn("postgres health check failed", zap.Error(err))
	} else {
		results["postgres"] = "healthy"
	}

	// Check S3
	if err := h.checkS3(ctx); err != nil {
		results["s3"] = "unhealthy: " + err.Error()
		allHealthy = false
		h.logger.Warn("s3 health check failed", zap.Error(err))
	} else {
		results["s3"] = "healthy"
	}

	// Check SNS (only if enabled)
	if h.awsConfig.SNS.Enabled {
		if err := h.checkSNS(ctx); err != nil {
			results["sns"] = "unhealthy: " + err.Error()
			allHealthy = false
			h.logger.Warn("sns health check failed", zap.Error(err))
		} else {
			results["sns"] = "healthy"
		}
	} else {
		results["sns"] = "disabled"
	}

	response := map[string]any{
		"checks":    results,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if allHealthy {
		response["status"] = "ready"
		return c.JSON(200, response)
	}

	response["status"] = "not ready"
	return c.JSON(503, response)
}

// Startup returns ECS startup health check
// Used during initial container startup with longer timeout
// Allows cold starts and initial connection establishment
func (h *Handler) Startup(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	results := make(map[string]string)
	allHealthy := true

	// Check all dependencies with more lenient timeout
	if err := h.checkPostgres(ctx); err != nil {
		results["postgres"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		results["postgres"] = "healthy"
	}

	if err := h.checkS3(ctx); err != nil {
		results["s3"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		results["s3"] = "healthy"
	}

	if h.awsConfig.SNS.Enabled {
		if err := h.checkSNS(ctx); err != nil {
			results["sns"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			results["sns"] = "healthy"
		}
	} else {
		results["sns"] = "disabled"
	}

	response := map[string]any{
		"checks":    results,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if allHealthy {
		response["status"] = "started"
		return c.JSON(200, response)
	}

	response["status"] = "starting"
	return c.JSON(503, response)
}

// checkPostgres verifies Postgres connectivity
func (h *Handler) checkPostgres(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return h.runner.Pool().Ping(checkCtx)
}

// checkS3 verifies S3 connectivity and bucket access
func (h *Handler) checkS3(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// HeadBucket is a lightweight check that verifies:
	// 1. S3 service is reachable
	// 2. Bucket exists
	// 3. We have permissions to access it
	_, err := h.s3Client.HeadBucket(checkCtx, &s3.HeadBucketInput{
		Bucket: &h.awsConfig.S3.Bucket,
	})

	return err
}

// checkSNS verifies SNS connectivity and topic access
func (h *Handler) checkSNS(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// If no topic ARN configured, SNS check passes
	// (service is configured to bypass SNS)
	if h.awsConfig.SNS.TopicARN == "" {
		return nil
	}

	// GetTopicAttributes is a lightweight check that verifies:
	// 1. SNS service is reachable
	// 2. Topic exists
	// 3. We have permissions to access it
	_, err := h.snsClient.GetTopicAttributes(checkCtx, &sns.GetTopicAttributesInput{
		TopicArn: &h.awsConfig.SNS.TopicARN,
	})

	return err
}
