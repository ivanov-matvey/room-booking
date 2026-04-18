package slothandler

import (
	"time"

	"github.com/google/uuid"
)

// SlotResponse is the API representation of a slot.
type SlotResponse struct {
	ID     uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	RoomID uuid.UUID `json:"roomId" swaggertype:"string" format:"uuid"`
	Start  time.Time `json:"start" swaggertype:"string" format:"date-time"`
	End    time.Time `json:"end" swaggertype:"string" format:"date-time"`
} //@name Slot

// ListSlotsResponse is the response body for ListSlots.
type ListSlotsResponse struct {
	Slots []SlotResponse `json:"slots"`
} //@name ListSlotsResponse
