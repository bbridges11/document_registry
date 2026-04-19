package document

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/bbridges_11/document-registry/internal/domain/document"
	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/domain/version"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/platform/storage"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type CommandService struct {
	documentRepo     outbound.DocumentRepository
	versionRepo      outbound.VersionRepository
	storageService   outbound.StorageService
	txManager        outbound.TransactionManager
	authz            outbound.AuthorizationService
	eventBus         outbound.EventBus
	contentValidator outbound.ContentValidator
}

func NewCommandService(
	documentRepo outbound.DocumentRepository,
	versionRepo outbound.VersionRepository,
	storageService outbound.StorageService,
	txManager outbound.TransactionManager,
	authz outbound.AuthorizationService,
	eventBus outbound.EventBus,
	contentValidator outbound.ContentValidator,
) *CommandService {
	return &CommandService{
		documentRepo:     documentRepo,
		versionRepo:      versionRepo,
		storageService:   storageService,
		txManager:        txManager,
		authz:            authz,
		eventBus:         eventBus,
		contentValidator: contentValidator,
	}
}

func (s *CommandService) CreateDocumentWithVersion(ctx context.Context, cmd CreateDocumentWithVersionCommand) (doc *document.Document, ver *version.Version, valResult *outbound.ValidationResult, err error) {
	defer err2.Handle(&err)

	// Step 1: Read content into memory with size limit
	const maxContentSize = 10 * 1024 * 1024 // 10MB
	limitReader := io.LimitReader(cmd.Content, maxContentSize+1)
	contentBytes := try.To1(io.ReadAll(limitReader))

	if len(contentBytes) > maxContentSize {
		return nil, nil, nil, errors.New(
			errors.CodeInvalidArgument,
			fmt.Sprintf("content exceeds maximum size of %d bytes", maxContentSize),
		)
	}

	// Step 2: Calculate content hash
	contentHash := storage.CalculateContentHash(contentBytes)

	// Step 3: VALIDATE CONTENT FIRST (before S3, before DB)
	var validationResult *outbound.ValidationResult
	if s.contentValidator.ShouldValidate(cmd.DocumentType) {
		validationResult = try.To1(s.contentValidator.Validate(ctx, outbound.ValidationRequest{
			DocumentType: cmd.DocumentType,
			Content:      contentBytes,
		}))

		// If validation failed, return immediately - NO S3 upload, NO DB transaction
		if !validationResult.Valid {
			return nil, nil, validationResult, errors.New(
				errors.CodeValidationFailed,
				"content validation failed",
			)
		}
	}

	// Step 4: Create document aggregate (generates UUID)
	doc = try.To1(document.NewDocument(cmd.Name, cmd.Description, cmd.DocumentType, cmd.CreatedBy, cmd.Tags))

	// Step 5: Generate S3 key using new format: {documentType}/{documentID}/{version}/content
	s3Key := storage.GenerateS3Key(cmd.DocumentType, doc.ID().String(), cmd.Version)

	// Step 6: Upload validated content to S3
	try.To(s.storageService.Upload(ctx, s3Key, bytes.NewReader(contentBytes)))

	// Step 7: Atomic DB transaction
	err = s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Save document
		try.To(s.documentRepo.Save(ctx, doc))

		// Create version aggregate with calculated hash
		ver = try.To1(version.NewVersion(doc.ID(), cmd.Version, s3Key, contentHash, cmd.CreatedBy, cmd.Metadata))

		// Save version
		try.To(s.versionRepo.Save(ctx, ver))

		return nil
	})

	// Step 8: Cleanup S3 on transaction failure
	if err != nil {
		_ = s.storageService.Delete(ctx, s3Key)
		return nil, nil, nil, errors.Wrap(errors.CodeInternal, "failed to create document with version", err)
	}

	// Step 9: Publish events for both aggregates
	s.publishDocumentEvents(ctx, doc, "")
	s.publishVersionEvents(ctx, ver, "")

	// Step 10: Publish validation audit event (if validation occurred)
	if validationResult != nil {
		s.publishValidationEvent(ctx, doc.ID().String(), ver.ID().String(), cmd.DocumentType, true, 0, cmd.CreatedBy)
	}

	return doc, ver, validationResult, nil
}

func (s *CommandService) UpdateDocument(ctx context.Context, cmd UpdateDocumentCommand) (err error) {
	defer err2.Handle(&err)

	// Check authorization
	allowed := try.To1(s.authz.CanAccessDocument(ctx, cmd.UpdatedBy, cmd.ID.String()))
	if !allowed {
		return errors.ErrForbidden
	}

	doc := try.To1(s.documentRepo.GetByID(ctx, cmd.ID))
	try.To(doc.Update(cmd.Name, cmd.Description, cmd.Tags, cmd.UpdatedBy))
	try.To(s.documentRepo.Save(ctx, doc))

	// Publish events after successful persistence
	s.publishDocumentEvents(ctx, doc, "")

	return nil
}

func (s *CommandService) publishDocumentEvents(ctx context.Context, doc *document.Document, traceID string) {
	domainEvents := doc.Events()
	if len(domainEvents) == 0 {
		return
	}

	envelopes := make([]events.Envelope, 0, len(domainEvents))
	for _, event := range domainEvents {
		envelopes = append(envelopes, events.NewEnvelope(event, traceID))
	}

	s.eventBus.Publish(ctx, envelopes...)
	doc.ClearEvents()
}

func (s *CommandService) publishVersionEvents(ctx context.Context, ver shared.EventEmitter, traceID string) {
	domainEvents := ver.Events()
	if len(domainEvents) == 0 {
		return
	}

	envelopes := make([]events.Envelope, 0, len(domainEvents))
	for _, event := range domainEvents {
		envelopes = append(envelopes, events.NewEnvelope(event, traceID))
	}

	s.eventBus.Publish(ctx, envelopes...)
	ver.ClearEvents()
}

func (s *CommandService) publishValidationEvent(ctx context.Context, documentID, versionID, documentType string, valid bool, issueCount int, validatedBy string) {
	envelope := events.NewEnvelope(
		events.NewValidationCompleted(documentID, versionID, documentType, valid, issueCount, validatedBy),
		"",
	)
	s.eventBus.Publish(ctx, envelope)
}
