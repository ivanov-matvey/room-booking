package response

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ivanov-matvey/room-booking/internal/domain"
)

type ErrorDetail struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"invalid request"`
} //@name ErrorDetail

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
} //@name ErrorResponse

type InternalErrorResponse struct {
	Error ErrorDetail `json:"error"`
} //@name InternalErrorResponse

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func DomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrRoomNotFound):
		Error(w, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrSlotNotFound):
		Error(w, http.StatusNotFound, "SLOT_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrBookingNotFound):
		Error(w, http.StatusNotFound, "BOOKING_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrSlotBooked):
		Error(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", err.Error())
	case errors.Is(err, domain.ErrScheduleExists):
		Error(w, http.StatusConflict, "SCHEDULE_EXISTS", err.Error())
	case errors.Is(err, domain.ErrForbidden):
		Error(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case errors.Is(err, domain.ErrPastSlot):
		Error(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	case errors.Is(err, domain.ErrInvalidRequest):
		Error(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	case errors.Is(err, domain.ErrEmailTaken):
		Error(w, http.StatusBadRequest, "EMAIL_TAKEN", err.Error())
	default:
		Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
