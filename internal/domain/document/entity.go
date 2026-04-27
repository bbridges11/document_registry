package document

import (
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
)

type Document struct {
	shared.AggregateRoot
	id           uuid.UUID
	name         string
	description  string
	tags         []string
	documentType DocumentType
	createdAt    time.Time
	updatedAt    time.Time
	createdBy    string
}

func NewDocument(name, description string, documentType DocumentType, createdBy string, tags []string) (*Document, error) {
	if name == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "name is required")
	}
	if createdBy == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "created by is required")
	}

	now := time.Now().UTC()
	doc := &Document{
		id:           uuid.New(),
		name:         name,
		description:  description,
		tags:         tags,
		documentType: documentType,
		createdAt:    now,
		updatedAt:    now,
		createdBy:    createdBy,
	}

	doc.AggregateRoot.AddEvent(events.NewDocumentCreated(
		doc.id.String(),
		doc.name,
		doc.documentType.Code(),
		doc.createdBy,
	))

	return doc, nil
}

func RehydrateDocument(id uuid.UUID, name, description string, documentType DocumentType, createdBy string, tags []string, createdAt, updatedAt time.Time) *Document {
	return &Document{
		id:           id,
		name:         name,
		description:  description,
		tags:         tags,
		documentType: documentType,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		createdBy:    createdBy,
	}
}

func (d *Document) Update(name, description string, tags []string, updatedBy string) error {
	if name == "" {
		return errors.New(errors.CodeInvalidArgument, "name is required")
	}

	d.name = name
	d.description = description
	d.tags = tags
	d.updatedAt = time.Now().UTC()

	d.AggregateRoot.AddEvent(events.NewDocumentUpdated(d.id.String(), updatedBy))

	return nil
}

func (d *Document) ID() uuid.UUID {
	return d.id
}

func (d *Document) Name() string {
	return d.name
}

func (d *Document) Description() string {
	return d.description
}

func (d *Document) Tags() []string {
	return d.tags
}

func (d *Document) DocumentType() DocumentType {
	return d.documentType
}

func (d *Document) CreatedAt() time.Time {
	return d.createdAt
}

func (d *Document) UpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Document) CreatedBy() string {
	return d.createdBy
}
