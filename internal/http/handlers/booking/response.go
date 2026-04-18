package bookinghandler

import (
	"time"

	"github.com/google/uuid"
)

// BookingResponse is the API representation of a booking.
type BookingResponse struct {
	ID             uuid.UUID  `json:"id" swaggertype:"string" format:"uuid"`
	SlotID         uuid.UUID  `json:"slotId" swaggertype:"string" format:"uuid"`
	UserID         uuid.UUID  `json:"userId" swaggertype:"string" format:"uuid"`
	Status         string     `json:"status" enums:"active,cancelled"`
	ConferenceLink *string    `json:"conferenceLink,omitempty" format:"uri"`
	CreatedAt      *time.Time `json:"createdAt,omitempty" swaggertype:"string" format:"date-time"`
} //@name Booking

// PaginationResponse is the API representation of pagination info.
type PaginationResponse struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
} //@name Pagination

// CreateBookingResponse is the response body for CreateBooking.
type CreateBookingResponse struct {
	Booking *BookingResponse `json:"booking"`
} //@name CreateBookingResponse

// ListBookingsResponse is the response body for ListBookings.
type ListBookingsResponse struct {
	Bookings   []BookingResponse   `json:"bookings"`
	Pagination *PaginationResponse `json:"pagination"`
} //@name ListBookingsResponse

// GetMyBookingsResponse is the response body for GetMyBookings.
type GetMyBookingsResponse struct {
	Bookings []BookingResponse `json:"bookings"`
} //@name GetMyBookingsResponse

// CancelBookingResponse is the response body for CancelBooking.
type CancelBookingResponse struct {
	Booking *BookingResponse `json:"booking"`
} //@name CancelBookingResponse
