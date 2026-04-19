package events

import "time"

// UserCreated event
type UserCreated struct {
	UserID     string
	Email      string
	Name       string
	Role       string
	occurredAt time.Time
}

func (e UserCreated) OccurredAt() time.Time {
	return e.occurredAt
}

func NewUserCreated(userID, email, name, role string) UserCreated {
	return UserCreated{
		UserID:     userID,
		Email:      email,
		Name:       name,
		Role:       role,
		occurredAt: time.Now().UTC(),
	}
}

func (e UserCreated) Type() EventType {
	return EventUserCreated
}

// UserUpdated event
type UserUpdated struct {
	UserID     string
	Name       string
	Role       string
	occurredAt time.Time
}

func NewUserUpdated(userID, name, role string) UserUpdated {
	return UserUpdated{
		UserID:     userID,
		Name:       name,
		Role:       role,
		occurredAt: time.Now().UTC(),
	}
}

func (e UserUpdated) Type() EventType {
	return EventUserUpdated
}

func (e UserUpdated) OccurredAt() time.Time {
	return e.occurredAt
}
