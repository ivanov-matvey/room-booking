package bookinghandler

import "github.com/google/uuid"

// CreateBookingRequest is the request body for CreateBooking.
type CreateBookingRequest struct {
	SlotID               uuid.UUID `json:"slotId" swaggertype:"string" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
	CreateConferenceLink *bool     `json:"createConferenceLink,omitempty" example:"false"`
} //@name CreateBookingRequest
