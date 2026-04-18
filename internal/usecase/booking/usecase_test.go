package booking_test

import (
	"context"
	"testing"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	bookingusecase "github.com/ivanov-matvey/room-booking/internal/usecase/booking"
	mockbooking "github.com/ivanov-matvey/room-booking/mocks/usecase/booking"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateBooking_Success(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	slotID := uuid.New()
	futureTime := time.Now().Add(time.Hour)

	slot := &domain.Slot{
		ID:        slotID,
		RoomID:    uuid.New(),
		StartTime: futureTime,
		EndTime:   futureTime.Add(30 * time.Minute),
	}

	slotRepo.On("GetByID", mock.Anything, slotID).Return(slot, nil)
	bookingRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Booking")).Return(nil)

	booking, err := uc.CreateBooking(context.Background(), userID, slotID, false)

	require.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, slotID, booking.SlotID)
	assert.Equal(t, userID, booking.UserID)
	assert.Equal(t, "active", booking.Status)
	assert.Nil(t, booking.ConferenceLink)

	slotRepo.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
}

func TestCreateBooking_WithConferenceLink(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)
	confService := mockbooking.NewMockConferenceService(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, confService)

	userID := uuid.New()
	slotID := uuid.New()
	futureTime := time.Now().Add(time.Hour)

	slot := &domain.Slot{
		ID:        slotID,
		RoomID:    uuid.New(),
		StartTime: futureTime,
		EndTime:   futureTime.Add(30 * time.Minute),
	}

	slotRepo.On("GetByID", mock.Anything, slotID).Return(slot, nil)
	bookingRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Booking")).Return(nil)
	confService.On("CreateConferenceLink", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return("https://meet.example.com/test", nil)

	booking, err := uc.CreateBooking(context.Background(), userID, slotID, true)

	require.NoError(t, err)
	assert.NotNil(t, booking)
	assert.NotNil(t, booking.ConferenceLink)

	slotRepo.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
	confService.AssertExpectations(t)
}

func TestCreateBooking_PastSlot(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	slotID := uuid.New()
	pastTime := time.Now().Add(-time.Hour)

	slot := &domain.Slot{
		ID:        slotID,
		RoomID:    uuid.New(),
		StartTime: pastTime,
		EndTime:   pastTime.Add(30 * time.Minute),
	}

	slotRepo.On("GetByID", mock.Anything, slotID).Return(slot, nil)

	booking, err := uc.CreateBooking(context.Background(), userID, slotID, false)

	require.Error(t, err)
	assert.Equal(t, domain.ErrPastSlot, err)
	assert.Nil(t, booking)
}

func TestCreateBooking_SlotNotFound(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	slotID := uuid.New()

	slotRepo.On("GetByID", mock.Anything, slotID).Return(nil, domain.ErrSlotNotFound)

	booking, err := uc.CreateBooking(context.Background(), userID, slotID, false)

	require.Error(t, err)
	assert.Equal(t, domain.ErrSlotNotFound, err)
	assert.Nil(t, booking)
}

func TestCreateBooking_AlreadyBooked(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	slotID := uuid.New()
	futureTime := time.Now().Add(time.Hour)

	slot := &domain.Slot{
		ID:        slotID,
		RoomID:    uuid.New(),
		StartTime: futureTime,
		EndTime:   futureTime.Add(30 * time.Minute),
	}

	slotRepo.On("GetByID", mock.Anything, slotID).Return(slot, nil)
	bookingRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Booking")).Return(domain.ErrSlotBooked)

	booking, err := uc.CreateBooking(context.Background(), userID, slotID, false)

	require.Error(t, err)
	assert.Equal(t, domain.ErrSlotBooked, err)
	assert.Nil(t, booking)
}

func TestCancelBooking_Success(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	bookingID := uuid.New()

	existingBooking := &domain.Booking{
		ID:     bookingID,
		SlotID: uuid.New(),
		UserID: userID,
		Status: "active",
	}

	bookingRepo.On("GetByID", mock.Anything, bookingID).Return(existingBooking, nil)
	bookingRepo.On("Cancel", mock.Anything, bookingID).Return(nil)

	booking, err := uc.CancelBooking(context.Background(), userID, bookingID)

	require.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, "cancelled", booking.Status)

	bookingRepo.AssertExpectations(t)
}

func TestCancelBooking_Idempotent(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	bookingID := uuid.New()

	existingBooking := &domain.Booking{
		ID:     bookingID,
		SlotID: uuid.New(),
		UserID: userID,
		Status: "cancelled",
	}

	bookingRepo.On("GetByID", mock.Anything, bookingID).Return(existingBooking, nil)

	booking, err := uc.CancelBooking(context.Background(), userID, bookingID)

	require.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, "cancelled", booking.Status)

	bookingRepo.AssertNotCalled(t, "Cancel")
}

func TestCancelBooking_Forbidden(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	otherUserID := uuid.New()
	bookingID := uuid.New()

	existingBooking := &domain.Booking{
		ID:     bookingID,
		SlotID: uuid.New(),
		UserID: otherUserID,
		Status: "active",
	}

	bookingRepo.On("GetByID", mock.Anything, bookingID).Return(existingBooking, nil)

	booking, err := uc.CancelBooking(context.Background(), userID, bookingID)

	require.Error(t, err)
	assert.Equal(t, domain.ErrForbidden, err)
	assert.Nil(t, booking)
}

func TestCancelBooking_NotFound(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	bookingID := uuid.New()

	bookingRepo.On("GetByID", mock.Anything, bookingID).Return(nil, domain.ErrBookingNotFound)

	booking, err := uc.CancelBooking(context.Background(), userID, bookingID)

	require.Error(t, err)
	assert.Equal(t, domain.ErrBookingNotFound, err)
	assert.Nil(t, booking)
}

func TestGetMyBookings_Success(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	bookings := []domain.Booking{
		{ID: uuid.New(), SlotID: uuid.New(), UserID: userID, Status: "active"},
		{ID: uuid.New(), SlotID: uuid.New(), UserID: userID, Status: "active"},
	}

	bookingRepo.On("ListByUserFuture", mock.Anything, userID).Return(bookings, nil)

	result, err := uc.GetMyBookings(context.Background(), userID)

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetMyBookings_Empty(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()

	bookingRepo.On("ListByUserFuture", mock.Anything, userID).Return([]domain.Booking{}, nil)

	result, err := uc.GetMyBookings(context.Background(), userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestListBookings_Success(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	bookings := []domain.Booking{
		{ID: uuid.New(), SlotID: uuid.New(), UserID: uuid.New(), Status: "active"},
	}

	bookingRepo.On("ListAll", mock.Anything, 1, 20).Return(bookings, 1, nil)

	result, total, err := uc.ListBookings(context.Background(), 1, 20)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}

func TestListBookings_PageSizeClamped(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	bookingRepo.On("ListAll", mock.Anything, 1, 100).Return([]domain.Booking{}, 0, nil)

	result, total, err := uc.ListBookings(context.Background(), 1, 200)

	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.NotNil(t, result)
}

func TestListBookings_PageNormalized(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	bookingRepo.On("ListAll", mock.Anything, 1, 20).Return([]domain.Booking{}, 0, nil)

	result, total, err := uc.ListBookings(context.Background(), 0, 20)

	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.NotNil(t, result)
}

func TestListBookings_PageSizeNormalized(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	bookingRepo.On("ListAll", mock.Anything, 1, 20).Return([]domain.Booking{}, 0, nil)

	result, total, err := uc.ListBookings(context.Background(), 1, 0)

	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.NotNil(t, result)
}

func TestListBookings_Error(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	bookingRepo.On("ListAll", mock.Anything, 1, 20).Return(nil, 0, assert.AnError)

	result, total, err := uc.ListBookings(context.Background(), 1, 20)

	require.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, result)
}

func TestGetMyBookings_Error(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	bookingRepo.On("ListByUserFuture", mock.Anything, userID).Return(nil, assert.AnError)

	result, err := uc.GetMyBookings(context.Background(), userID)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestCreateBooking_ConferenceLinkFailed(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)
	confService := mockbooking.NewMockConferenceService(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, confService)

	userID := uuid.New()
	slotID := uuid.New()
	futureTime := time.Now().Add(time.Hour)

	slot := &domain.Slot{
		ID:        slotID,
		RoomID:    uuid.New(),
		StartTime: futureTime,
		EndTime:   futureTime.Add(30 * time.Minute),
	}

	slotRepo.On("GetByID", mock.Anything, slotID).Return(slot, nil)
	confService.On("CreateConferenceLink", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return("", assert.AnError)
	bookingRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Booking")).Return(nil)

	booking, err := uc.CreateBooking(context.Background(), userID, slotID, true)

	require.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Nil(t, booking.ConferenceLink)

	slotRepo.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
	confService.AssertExpectations(t)
}

func TestCancelBooking_CancelError(t *testing.T) {
	bookingRepo := mockbooking.NewMockRepository(t)
	slotRepo := mockbooking.NewMockSlotRepository(t)

	uc := bookingusecase.New(bookingRepo, slotRepo, nil)

	userID := uuid.New()
	bookingID := uuid.New()

	existingBooking := &domain.Booking{
		ID:     bookingID,
		SlotID: uuid.New(),
		UserID: userID,
		Status: "active",
	}

	bookingRepo.On("GetByID", mock.Anything, bookingID).Return(existingBooking, nil)
	bookingRepo.On("Cancel", mock.Anything, bookingID).Return(assert.AnError)

	booking, err := uc.CancelBooking(context.Background(), userID, bookingID)

	require.Error(t, err)
	assert.Nil(t, booking)
}
