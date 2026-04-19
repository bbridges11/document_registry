package bootstrap

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/approval"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/deprecation"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/document"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/health"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/stakeholder"
	validationhttp "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/validation"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/version"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/validation/inprocess"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	"github.com/bbridges_11/document-registry/internal/platform/config"
	"go.uber.org/zap"

	userhttp "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/user"
	appApproval "github.com/bbridges_11/document-registry/internal/application/approval"
	appDeprecation "github.com/bbridges_11/document-registry/internal/application/deprecation"
	appDocument "github.com/bbridges_11/document-registry/internal/application/document"
	appStakeholder "github.com/bbridges_11/document-registry/internal/application/stakeholder"
	appUser "github.com/bbridges_11/document-registry/internal/application/user"
	appValidation "github.com/bbridges_11/document-registry/internal/application/validation"
	appVersion "github.com/bbridges_11/document-registry/internal/application/version"
	domainApproval "github.com/bbridges_11/document-registry/internal/domain/approval"
)

type httpHandlers struct {
	Health      *health.Handler
	Document    *document.Handler
	Version     *version.Handler
	Stakeholder *stakeholder.Handler
	Approval    *approval.Handler
	Deprecation *deprecation.Handler
	User        *userhttp.Handler
	Validation  *validationhttp.Handler
	UserQueries appUser.QueryService
}

func wireApplication(_ context.Context, _ *config.Config, log *zap.Logger, infra *infrastructure) *httpHandlers {
	log.Info("initializing application services")
	workflowFactory := workflow.NewFactory()
	policyFactory := domainApproval.NewPolicyFactory()

	registerEventHandlers(infra, log)

	// Create content validator
	definitionValidator := inprocess.NewDefinitionValidator()

	documentCommandService := appDocument.NewCommandService(
		infra.DocumentRepo,
		infra.VersionRepo,
		infra.StorageService,
		infra.Runner,
		infra.AuthzService,
		infra.EventBus,
		definitionValidator,
	)
	documentQueryService := appDocument.NewQueryService(
		infra.DocumentRepo,
		infra.VersionRepo,
		infra.AuthzService,
	)
	documentService := appDocument.NewService(documentCommandService, documentQueryService)

	versionCommandService := appVersion.NewCommandService(
		infra.VersionRepo,
		infra.DocumentRepo,
		infra.ApprovalRepo,
		infra.DeprecationRepo,
		infra.StorageService,
		infra.Runner,
		infra.AuthzService,
		infra.UserRepo,
		infra.EventBus,
		workflowFactory,
		policyFactory,
		definitionValidator,
	)
	versionQueryService := appVersion.NewQueryService(
		infra.VersionRepo,
		infra.DocumentRepo,
		infra.ApprovalRepo,
		infra.AuthzService,
		infra.UserRepo,
		workflowFactory,
		policyFactory,
	)
	versionService := appVersion.NewService(versionCommandService, versionQueryService)

	stakeholderCommandService := appStakeholder.NewCommandService(
		infra.StakeholderRepo,
		infra.DocumentRepo,
		infra.AuthzService,
		infra.UserRepo,
		infra.EventBus,
	)
	stakeholderQueryService := appStakeholder.NewQueryService(
		infra.StakeholderRepo,
		infra.DocumentRepo,
		infra.AuthzService,
		infra.UserRepo,
	)
	stakeholderService := appStakeholder.NewService(stakeholderCommandService, stakeholderQueryService)

	approvalCommandService := appApproval.NewCommandService(
		infra.ApprovalRepo,
		infra.VersionRepo,
		infra.DocumentRepo,
		infra.UserRepo,
		policyFactory,
	)
	approvalQueryService := appApproval.NewQueryService(
		infra.ApprovalRepo,
		infra.VersionRepo,
		infra.DocumentRepo,
		infra.UserRepo,
		policyFactory,
	)
	approvalService := appApproval.NewService(approvalCommandService, approvalQueryService)

	deprecationCommandService := appDeprecation.NewCommandService(
		infra.DeprecationRepo,
		infra.VersionRepo,
		infra.Runner,
		infra.EventBus,
		infra.AuthzService,
	)
	deprecationService := appDeprecation.NewService(deprecationCommandService)

	userCommands := appUser.NewUserCommandService(infra.UserRepo, infra.EventBus, log)
	userQueries := appUser.NewUserQueryService(infra.UserRepo, log)

	// Create validation service
	validationService := appValidation.NewService(definitionValidator)

	return &httpHandlers{
		Health:      health.NewHandler(infra.Runner, infra.AuthzService, infra.UserService, log),
		Document:    document.NewHandler(documentService),
		Version:     version.NewHandler(versionService),
		Stakeholder: stakeholder.NewHandler(stakeholderService),
		Approval:    approval.NewHandler(approvalService),
		Deprecation: deprecation.NewHandler(deprecationService),
		User:        userhttp.NewHandler(userCommands, userQueries),
		Validation:  validationhttp.NewHandler(validationService),
		UserQueries: userQueries,
	}
}
