package bootstrap

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/adapters/outbound/authorization/openfga"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/persistence/postgres"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/storage/s3"
	useradapter "github.com/bbridges_11/document-registry/internal/adapters/outbound/user"
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
	Runner          *dbPostgres.Runner
	EventBus        *platformEvents.Bus
	DocumentRepo    outbound.DocumentRepository
	VersionRepo     outbound.VersionRepository
	StakeholderRepo outbound.StakeholderRepository
	ApprovalRepo    outbound.ApprovalRepository
	UserRepo        outbound.UserRepository
	DeprecationRepo outbound.DeprecationRepository
	StorageService  outbound.StorageService
	AuthzService    outbound.AuthorizationService
	UserService     outbound.UserService
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
		pool := try.To1(dbPostgres.NewPool(ctx, cfg.Database.Postgres))
		runner = dbPostgres.NewRunner(pool)
		log.Info("postgres pool initialized")
	}

	s3Client := try.To1(aws.NewS3Client(ctx, cfg.AWS))
	log.Info("s3 client initialized")

	eventBus := platformEvents.NewBus(log, 1000, 4)
	eventBus.Use(platformEvents.RecoveryMiddleware(log))
	eventBus.Use(platformEvents.LoggingMiddleware(log))
	try.To(eventBus.Start(ctx))
	log.Info("event bus started")

	authzService := try.To1(openfga.NewAuthorizationAdapter(cfg.OpenFGA, log))

	return &infrastructure{
		Runner:          runner,
		EventBus:        eventBus,
		DocumentRepo:    postgres.NewDocumentRepository(runner),
		VersionRepo:     postgres.NewVersionRepository(runner),
		StakeholderRepo: postgres.NewStakeholderRepository(runner),
		ApprovalRepo:    postgres.NewApprovalRepository(runner),
		UserRepo:        postgres.NewUserRepository(runner),
		DeprecationRepo: postgres.NewDeprecationRepository(runner),
		StorageService:  s3.NewStorageAdapter(s3Client, cfg.AWS.S3),
		AuthzService:    authzService,
		UserService:     useradapter.NewServiceAdapter(cfg.User),
	}
}
