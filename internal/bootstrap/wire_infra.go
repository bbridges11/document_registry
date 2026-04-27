package bootstrap

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	notificationSNS "github.com/bbridges_11/document-registry/internal/adapters/outbound/notification/sns"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/persistence/postgres"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/publisher/mock"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/storage"
	storageS3 "github.com/bbridges_11/document-registry/internal/adapters/outbound/storage/s3"
	"github.com/bbridges_11/document-registry/internal/platform/aws"
	"github.com/bbridges_11/document-registry/internal/platform/config"
	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	platformEvents "github.com/bbridges_11/document-registry/internal/platform/events"
	"github.com/bbridges_11/document-registry/internal/platform/logger"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/lainio/err2/try"
	"go.uber.org/zap"
)

type infrastructure struct {
	Runner              *dbPostgres.Runner
	EventBus            *platformEvents.Bus
	DocumentRepo        outbound.DocumentRepository
	VersionRepo         outbound.VersionRepository
	StakeholderRepo     outbound.StakeholderRepository
	ApprovalRepo        outbound.ApprovalRepository
	UserRepo            outbound.UserRepository
	DeprecationRepo     outbound.DeprecationRepository
	PublicationRepo     outbound.PublicationRepository
	StorageService      outbound.StorageService
	PublisherService    outbound.PublisherService
	NotificationService outbound.NotificationService
	S3Client            *s3.Client
	SNSClient           *sns.Client
	AWSConfig           config.AWSConfig
}

func wireConfigAndLogger() (*config.Config, *zap.Logger) {
	cfg := try.To1(config.Load())
	log := try.To1(logger.New(cfg.Logging))
	log.Info("initializing application",
		zap.String("app", cfg.App.Name),
		zap.String("version", cfg.App.Version),
		zap.String("profile", string(cfg.App.Profile)),
	)
	return cfg, log
}

func wireInfrastructure(ctx context.Context, cfg *config.Config, log *zap.Logger) *infrastructure {
	log.Info("initializing platform infrastructure")

	var runner *dbPostgres.Runner
	if cfg.Database.Driver == config.DatabaseDriverPostgres {
		// Pass profile and AWS region to pool creation
		pool := try.To1(dbPostgres.NewPool(
			ctx,
			cfg.Database.Postgres,
			cfg.App.Profile,
			cfg.AWS.Region,
		))
		runner = dbPostgres.NewRunner(pool)

		// Log connection method
		if cfg.Database.Postgres.IsCloudProfile(cfg.App.Profile) {
			log.Info("postgres pool initialized with IAM authentication",
				zap.String("host", cfg.Database.Postgres.Host),
				zap.Int("port", cfg.Database.Postgres.Port),
				zap.String("database", cfg.Database.Postgres.Database),
				zap.String("user", cfg.Database.Postgres.User),
				zap.String("region", cfg.AWS.Region),
			)
		} else {
			log.Info("postgres pool initialized with password authentication",
				zap.String("host", cfg.Database.Postgres.Host),
				zap.Int("port", cfg.Database.Postgres.Port),
				zap.String("database", cfg.Database.Postgres.Database),
			)
		}
	}

	s3Client := try.To1(aws.NewS3Client(ctx, cfg.AWS))
	log.Info("s3 client initialized")

	snsClient := try.To1(aws.NewSNSClient(ctx, cfg.AWS))
	log.Info("SNS client initialized")

	eventBus := platformEvents.NewBus(log, 1000, 4)
	eventBus.Use(platformEvents.RecoveryMiddleware(log))
	eventBus.Use(platformEvents.LoggingMiddleware(log))
	try.To(eventBus.Start(ctx))
	log.Info("event bus started")

	publisherService := mock.NewPublisherAdapter(log)

	return &infrastructure{
		Runner:              runner,
		EventBus:            eventBus,
		DocumentRepo:        postgres.NewDocumentRepository(runner),
		VersionRepo:         postgres.NewVersionRepository(runner),
		StakeholderRepo:     postgres.NewStakeholderRepository(runner),
		ApprovalRepo:        postgres.NewApprovalRepository(runner),
		UserRepo:            postgres.NewUserRepository(runner),
		DeprecationRepo:     postgres.NewDeprecationRepository(runner),
		StorageService:      storage.NewRouter(storageS3.NewBackend(s3Client, cfg.AWS.S3)),
		NotificationService: notificationSNS.NewSNSAdapter(snsClient, cfg.AWS.SNS, log),
		PublisherService:    publisherService,
		PublicationRepo:     postgres.NewPublicationRepository(runner),
		S3Client:            s3Client,
		SNSClient:           snsClient,
		AWSConfig:           cfg.AWS,
	}
}
