package domain

import "errors"

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrSlotNotFound    = errors.New("slot not found")
	ErrBookingNotFound = errors.New("booking not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrSlotBooked      = errors.New("slot already booked")
	ErrScheduleExists  = errors.New("schedule already exists")
	ErrEmailTaken      = errors.New("email already taken")
	ErrForbidden       = errors.New("forbidden")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrPastSlot        = errors.New("slot is in the past")
	ErrInvalidRequest  = errors.New("invalid request")
)
