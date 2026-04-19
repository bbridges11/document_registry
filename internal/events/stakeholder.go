package events

import "time"

type StakeholderAdded struct {
	DocumentID string
	UserID     string
	Role       string
	AddedBy    string
	occurredAt time.Time
}

func NewStakeholderAdded(documentID, userID, role, addedBy string) StakeholderAdded {
	return StakeholderAdded{
		DocumentID: documentID,
		UserID:     userID,
		Role:       role,
		AddedBy:    addedBy,
		occurredAt: time.Now().UTC(),
	}
}

func (e StakeholderAdded) Type() EventType {
	return EventStakeholderAdded
}

func (e StakeholderAdded) OccurredAt() time.Time {
	return e.occurredAt
}

type StakeholderRemoved struct {
	DocumentID string
	UserID     string
	Role       string
	RemovedBy  string
	occurredAt time.Time
}

func NewStakeholderRemoved(documentID, userID, role, removedBy string) StakeholderRemoved {
	return StakeholderRemoved{
		DocumentID: documentID,
		UserID:     userID,
		Role:       role,
		RemovedBy:  removedBy,
		occurredAt: time.Now().UTC(),
	}
}

func (e StakeholderRemoved) Type() EventType {
	return EventStakeholderRemoved
}

func (e StakeholderRemoved) OccurredAt() time.Time {
	return e.occurredAt
}
