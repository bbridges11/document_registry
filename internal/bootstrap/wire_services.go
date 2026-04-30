package bootstrap

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/approval"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/deprecation"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/document"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/health"
	publicationhttp "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/publication"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/stakeholder"
	subscriptionhttp "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/subscription"
	validationhttp "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/validation"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/version"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/aws"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/persistence/postgres"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/validation"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/validation/inprocess"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	"github.com/bbridges_11/document-registry/internal/platform/config"
	"go.uber.org/zap"

	userhttp "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/user"
	appApproval "github.com/bbridges_11/document-registry/internal/application/approval"
	appDeprecation "github.com/bbridges_11/document-registry/internal/application/deprecation"
	appDocument "github.com/bbridges_11/document-registry/internal/application/document"
	appPublication "github.com/bbridges_11/document-registry/internal/application/publication"
	appStakeholder "github.com/bbridges_11/document-registry/internal/application/stakeholder"
	appSubscription "github.com/bbridges_11/document-registry/internal/application/subscription"
	appUser "github.com/bbridges_11/document-registry/internal/application/user"
	appValidation "github.com/bbridges_11/document-registry/internal/application/validation"
	appVersion "github.com/bbridges_11/document-registry/internal/application/version"
	domainApproval "github.com/bbridges_11/document-registry/internal/domain/approval"
)

type httpHandlers struct {
	Health       *health.Handler
	Document     *document.Handler
	Version      *version.Handler
	Stakeholder  *stakeholder.Handler
	Approval     *approval.Handler
	Deprecation  *deprecation.Handler
	Publication  *publicationhttp.Handler
	Subscription *subscriptionhttp.Handler
	User         *userhttp.Handler
	Validation   *validationhttp.Handler
	UserQueries  appUser.QueryService
}

func wireApplication(_ context.Context, _ *config.Config, log *zap.Logger, infra *infrastructure) *httpHandlers {
	log.Info("initializing application services")
	workflowFactory := workflow.NewFactory()
	policyFactory := domainApproval.NewPolicyFactory()

	registerEventHandlers(infra, log)

	// Create validation registry with validators
	patternValidator := inprocess.NewPatternValidator()
	validatorRegistry := validation.NewRegistry(patternValidator)
	contentValidator := inprocess.NewContentValidator(validatorRegistry)

	documentCommandService := appDocument.NewCommandService(
		infra.DocumentRepo,
		infra.VersionRepo,
		infra.StorageService,
		infra.Runner,
		infra.EventBus,
		contentValidator,
	)
	documentQueryService := appDocument.NewQueryService(
		infra.DocumentRepo,
		infra.VersionRepo,
	)
	documentService := appDocument.NewService(documentCommandService, documentQueryService)

	// Create UserRoleResolver for version approvals
	userRoleResolver := postgres.NewUserRoleResolverAdapter(infra.UserRepo)

	versionCommandService := appVersion.NewCommandService(
		infra.VersionRepo,
		infra.DocumentRepo,
		infra.ApprovalRepo,
		infra.DeprecationRepo,
		infra.PublicationRepo,
		infra.StorageService,
		infra.PublisherService,
		infra.Runner,
		infra.UserRepo,
		userRoleResolver,
		infra.EventBus,
		workflowFactory,
		policyFactory,
		contentValidator,
	)
	versionQueryService := appVersion.NewQueryService(
		infra.VersionRepo,
		infra.DocumentRepo,
		infra.ApprovalRepo,
		infra.UserRepo,
		workflowFactory,
		policyFactory,
	)
	versionService := appVersion.NewService(versionCommandService, versionQueryService)

	stakeholderCommandService := appStakeholder.NewCommandService(
		infra.StakeholderRepo,
		infra.DocumentRepo,
		infra.UserRepo,
		infra.EventBus,
	)
	stakeholderQueryService := appStakeholder.NewQueryService(
		infra.StakeholderRepo,
		infra.DocumentRepo,
		infra.UserRepo,
	)
	stakeholderService := appStakeholder.NewService(stakeholderCommandService, stakeholderQueryService)

	approvalQueryService := appApproval.NewQueryService(
		infra.ApprovalRepo,
		infra.VersionRepo,
		infra.DocumentRepo,
		infra.UserRepo,
		policyFactory,
	)
	approvalService := appApproval.NewService(approvalQueryService)

	deprecationCommandService := appDeprecation.NewCommandService(
		infra.DeprecationRepo,
		infra.VersionRepo,
		infra.Runner,
		infra.EventBus,
	)
	deprecationService := appDeprecation.NewService(deprecationCommandService)

	publicationQueryService := appPublication.NewQueryService(infra.PublicationRepo)

	// Create SNS service and subscription service
	snsService := aws.NewSNSServiceAdapter(infra.SNSClient, infra.AWSConfig.SNS)
	subscriptionService := appSubscription.NewService(infra.UserRepo, snsService)

	userCommands := appUser.NewUserCommandService(infra.UserRepo, infra.EventBus, log)
	userQueries := appUser.NewUserQueryService(infra.UserRepo, log)

	// Create validation service
	validationService := appValidation.NewService(contentValidator)

	return &httpHandlers{
		Health:       health.NewHandler(infra.Runner, infra.S3Client, infra.SNSClient, infra.AWSConfig, log),
		Document:     document.NewHandler(documentService),
		Version:      version.NewHandler(versionService),
		Stakeholder:  stakeholder.NewHandler(stakeholderService),
		Approval:     approval.NewHandler(approvalService),
		Deprecation:  deprecation.NewHandler(deprecationService),
		Publication:  publicationhttp.NewHandler(publicationQueryService),
		Subscription: subscriptionhttp.NewHandler(subscriptionService),
		User:         userhttp.NewHandler(userCommands, userQueries),
		Validation:   validationhttp.NewHandler(validationService),
		UserQueries:  userQueries,
	}
}
