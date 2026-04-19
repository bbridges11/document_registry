package events

import "time"

type DocumentCreated struct {
	DocumentID   string
	Name         string
	DocumentType string
	CreatedBy    string
	occurredAt   time.Time
}

func NewDocumentCreated(documentID, name, documentType, createdBy string) DocumentCreated {
	return DocumentCreated{
		DocumentID:   documentID,
		Name:         name,
		DocumentType: documentType,
		CreatedBy:    createdBy,
		occurredAt:   time.Now().UTC(),
	}
}

func (e DocumentCreated) Type() EventType {
	return EventDocumentCreated
}

func (e DocumentCreated) OccurredAt() time.Time {
	return e.occurredAt
}

type DocumentUpdated struct {
	DocumentID string
	UpdatedBy  string
	occurredAt time.Time
}

func NewDocumentUpdated(documentID, updatedBy string) DocumentUpdated {
	return DocumentUpdated{
		DocumentID: documentID,
		UpdatedBy:  updatedBy,
		occurredAt: time.Now().UTC(),
	}
}

func (e DocumentUpdated) Type() EventType {
	return EventDocumentUpdated
}

func (e DocumentUpdated) OccurredAt() time.Time {
	return e.occurredAt
}
