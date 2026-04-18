package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	Role         string
	PasswordHash *string
	CreatedAt    time.Time
}

type Room struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Capacity    *int
	CreatedAt   time.Time
}

type Schedule struct {
	ID         uuid.UUID
	RoomID     uuid.UUID
	DaysOfWeek []int
	StartTime  string
	EndTime    string
	CreatedAt  time.Time
}

type Slot struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	StartTime time.Time
	EndTime   time.Time
	CreatedAt time.Time
}

const (
	BookingStatusActive    = "active"
	BookingStatusCancelled = "cancelled"

	RoleAdmin = "admin"
	RoleUser  = "user"
)

type Booking struct {
	ID             uuid.UUID
	SlotID         uuid.UUID
	UserID         uuid.UUID
	Status         string
	ConferenceLink *string
	CreatedAt      time.Time
}
