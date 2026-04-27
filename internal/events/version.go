package events

import "time"

type VersionCreated struct {
	VersionID  string
	DocumentID string
	Version    string
	CreatedBy  string
	occurredAt time.Time
}

func NewVersionCreated(versionID, documentID, version, createdBy string) VersionCreated {
	return VersionCreated{
		VersionID:  versionID,
		DocumentID: documentID,
		Version:    version,
		CreatedBy:  createdBy,
		occurredAt: time.Now().UTC(),
	}
}

func (e VersionCreated) Type() EventType {
	return EventVersionCreated
}

func (e VersionCreated) OccurredAt() time.Time {
	return e.occurredAt
}

type VersionSubmitted struct {
	VersionID   string
	DocumentID  string
	Version     string
	SubmittedBy string
	occurredAt  time.Time
}

func NewVersionSubmitted(versionID, documentID, version, submittedBy string) VersionSubmitted {
	return VersionSubmitted{
		VersionID:   versionID,
		DocumentID:  documentID,
		Version:     version,
		SubmittedBy: submittedBy,
		occurredAt:  time.Now().UTC(),
	}
}

func (e VersionSubmitted) Type() EventType {
	return EventVersionSubmitted
}

func (e VersionSubmitted) OccurredAt() time.Time {
	return e.occurredAt
}

type VersionInReview struct {
	VersionID  string
	DocumentID string
	Version    string
	ReviewedBy string
	occurredAt time.Time
}

func NewVersionInReview(versionID, documentID, version, reviewedBy string) VersionInReview {
	return VersionInReview{
		VersionID:  versionID,
		DocumentID: documentID,
		Version:    version,
		ReviewedBy: reviewedBy,
		occurredAt: time.Now().UTC(),
	}
}

func (e VersionInReview) Type() EventType {
	return EventVersionInReview
}

func (e VersionInReview) OccurredAt() time.Time {
	return e.occurredAt
}

type VersionApproved struct {
	VersionID  string
	DocumentID string
	Version    string
	ApprovedBy string
	Role       string
	occurredAt time.Time
}

func NewVersionApproved(versionID, documentID, version, approvedBy, role string) VersionApproved {
	return VersionApproved{
		VersionID:  versionID,
		DocumentID: documentID,
		Version:    version,
		ApprovedBy: approvedBy,
		Role:       role,
		occurredAt: time.Now().UTC(),
	}
}

func (e VersionApproved) Type() EventType {
	return EventVersionApproved
}

func (e VersionApproved) OccurredAt() time.Time {
	return e.occurredAt
}

type VersionFullyApproved struct {
	VersionID  string
	DocumentID string
	Version    string
	occurredAt time.Time
}

func NewVersionFullyApproved(versionID, documentID, version string) VersionFullyApproved {
	return VersionFullyApproved{
		VersionID:  versionID,
		DocumentID: documentID,
		Version:    version,
		occurredAt: time.Now().UTC(),
	}
}

func (e VersionFullyApproved) Type() EventType {
	return EventVersionFullyApproved
}

func (e VersionFullyApproved) OccurredAt() time.Time {
	return e.occurredAt
}

type VersionRejected struct {
	VersionID  string
	DocumentID string
	Version    string
	RejectedBy string
	Reason     string
	occurredAt time.Time
}

func NewVersionRejected(versionID, documentID, version, rejectedBy, reason string) VersionRejected {
	return VersionRejected{
		VersionID:  versionID,
		DocumentID: documentID,
		Version:    version,
		RejectedBy: rejectedBy,
		Reason:     reason,
		occurredAt: time.Now().UTC(),
	}
}

func (e VersionRejected) Type() EventType {
	return EventVersionRejected
}

func (e VersionRejected) OccurredAt() time.Time {
	return e.occurredAt
}

type VersionPublished struct {
	VersionID   string
	DocumentID  string
	Version     string
	PublishedBy string
	occurredAt  time.Time
}

func NewVersionPublished(versionID, documentID, version, publishedBy string) VersionPublished {
	return VersionPublished{
		VersionID:   versionID,
		DocumentID:  documentID,
		Version:     version,
		PublishedBy: publishedBy,
		occurredAt:  time.Now().UTC(),
	}
}

func (e VersionPublished) Type() EventType {
	return EventVersionPublished
}

func (e VersionPublished) OccurredAt() time.Time {
	return e.occurredAt
}
