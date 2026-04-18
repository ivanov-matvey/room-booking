package bookinghandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	bookinghandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/booking"
	"github.com/ivanov-matvey/room-booking/internal/http/middleware"
	mockbookinghandler "github.com/ivanov-matvey/room-booking/mocks/http/handlers/booking"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func withChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextUserID, userID)
	return r.WithContext(ctx)
}

func TestCreateBooking_HappyPath(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()

	booking := &domain.Booking{
		ID:        bookingID,
		SlotID:    slotID,
		UserID:    userID,
		Status:    domain.BookingStatusActive,
		CreatedAt: now,
	}

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().CreateBooking(mock.Anything, userID, slotID, false).Return(booking, nil)

	h := bookinghandler.New(m)

	body, _ := json.Marshal(map[string]interface{}{"slotId": slotID.String()})
	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBuffer(body))
	req = withUserID(req, userID.String())
	w := httptest.NewRecorder()
	h.CreateBooking(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp bookinghandler.CreateBookingResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotNil(t, resp.Booking)
	assert.Equal(t, bookingID, resp.Booking.ID)
}

func TestCreateBooking_WithConferenceLink(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()
	link := "https://meet.example.com/room123"

	booking := &domain.Booking{
		ID:             bookingID,
		SlotID:         slotID,
		UserID:         userID,
		Status:         domain.BookingStatusActive,
		ConferenceLink: &link,
		CreatedAt:      now,
	}

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().CreateBooking(mock.Anything, userID, slotID, true).Return(booking, nil)

	h := bookinghandler.New(m)

	createLink := true
	body, _ := json.Marshal(map[string]interface{}{"slotId": slotID.String(), "createConferenceLink": createLink})
	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBuffer(body))
	req = withUserID(req, userID.String())
	w := httptest.NewRecorder()
	h.CreateBooking(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp bookinghandler.CreateBookingResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotNil(t, resp.Booking)
	require.NotNil(t, resp.Booking.ConferenceLink)
	assert.Equal(t, link, *resp.Booking.ConferenceLink)
}

func TestCreateBooking_MissingUserID(t *testing.T) {
	m := mockbookinghandler.NewMockUseCase(t)

	h := bookinghandler.New(m)

	slotID := uuid.New()
	body, _ := json.Marshal(map[string]interface{}{"slotId": slotID.String()})
	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	h.CreateBooking(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateBooking_InvalidJSON(t *testing.T) {
	userID := uuid.New()
	m := mockbookinghandler.NewMockUseCase(t)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBufferString("not-json"))
	req = withUserID(req, userID.String())
	w := httptest.NewRecorder()
	h.CreateBooking(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBooking_SlotBooked(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().CreateBooking(mock.Anything, userID, slotID, false).Return(nil, domain.ErrSlotBooked)

	h := bookinghandler.New(m)

	body, _ := json.Marshal(map[string]interface{}{"slotId": slotID.String()})
	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBuffer(body))
	req = withUserID(req, userID.String())
	w := httptest.NewRecorder()
	h.CreateBooking(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateBooking_SlotNotFound(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().CreateBooking(mock.Anything, userID, slotID, false).Return(nil, domain.ErrSlotNotFound)

	h := bookinghandler.New(m)

	body, _ := json.Marshal(map[string]interface{}{"slotId": slotID.String()})
	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBuffer(body))
	req = withUserID(req, userID.String())
	w := httptest.NewRecorder()
	h.CreateBooking(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateBooking_InternalError(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().CreateBooking(mock.Anything, userID, slotID, false).Return(nil, assert.AnError)

	h := bookinghandler.New(m)

	body, _ := json.Marshal(map[string]interface{}{"slotId": slotID.String()})
	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBuffer(body))
	req = withUserID(req, userID.String())
	w := httptest.NewRecorder()
	h.CreateBooking(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListBookings_HappyPath(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()
	now := time.Now()

	bookings := []domain.Booking{
		{
			ID:        uuid.New(),
			SlotID:    slotID,
			UserID:    userID,
			Status:    domain.BookingStatusActive,
			CreatedAt: now,
		},
	}

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().ListBookings(mock.Anything, 1, 20).Return(bookings, 1, nil)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/bookings/list", nil)
	w := httptest.NewRecorder()
	h.ListBookings(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp bookinghandler.ListBookingsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Len(t, resp.Bookings, 1)
	require.NotNil(t, resp.Pagination)
	assert.Equal(t, 1, resp.Pagination.Total)
}

func TestListBookings_WithPagination(t *testing.T) {
	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().ListBookings(mock.Anything, 2, 10).Return([]domain.Booking{}, 0, nil)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/bookings/list?page=2&pageSize=10", nil)
	w := httptest.NewRecorder()
	h.ListBookings(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListBookings_InvalidPage(t *testing.T) {
	m := mockbookinghandler.NewMockUseCase(t)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/bookings/list?page=0", nil)
	w := httptest.NewRecorder()
	h.ListBookings(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListBookings_InvalidPageSize(t *testing.T) {
	m := mockbookinghandler.NewMockUseCase(t)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/bookings/list?pageSize=200", nil)
	w := httptest.NewRecorder()
	h.ListBookings(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListBookings_InternalError(t *testing.T) {
	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().ListBookings(mock.Anything, 1, 20).Return(nil, 0, assert.AnError)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/bookings/list", nil)
	w := httptest.NewRecorder()
	h.ListBookings(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetMyBookings_HappyPath(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()
	now := time.Now()

	bookings := []domain.Booking{
		{
			ID:        uuid.New(),
			SlotID:    slotID,
			UserID:    userID,
			Status:    domain.BookingStatusActive,
			CreatedAt: now,
		},
	}

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().GetMyBookings(mock.Anything, userID).Return(bookings, nil)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
	req = withUserID(req, userID.String())
	w := httptest.NewRecorder()
	h.GetMyBookings(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp bookinghandler.GetMyBookingsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Len(t, resp.Bookings, 1)
}

func TestGetMyBookings_MissingUserID(t *testing.T) {
	m := mockbookinghandler.NewMockUseCase(t)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
	w := httptest.NewRecorder()
	h.GetMyBookings(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMyBookings_InternalError(t *testing.T) {
	userID := uuid.New()

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().GetMyBookings(mock.Anything, userID).Return(nil, assert.AnError)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
	req = withUserID(req, userID.String())
	w := httptest.NewRecorder()
	h.GetMyBookings(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCancelBooking_HappyPath(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	slotID := uuid.New()
	now := time.Now()

	booking := &domain.Booking{
		ID:        bookingID,
		SlotID:    slotID,
		UserID:    userID,
		Status:    domain.BookingStatusCancelled,
		CreatedAt: now,
	}

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().CancelBooking(mock.Anything, userID, bookingID).Return(booking, nil)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/cancel", nil)
	req = withUserID(req, userID.String())
	req = withChiParams(req, map[string]string{"bookingId": bookingID.String()})
	w := httptest.NewRecorder()
	h.CancelBooking(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp bookinghandler.CancelBookingResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotNil(t, resp.Booking)
	assert.Equal(t, domain.BookingStatusCancelled, resp.Booking.Status)
}

func TestCancelBooking_MissingUserID(t *testing.T) {
	bookingID := uuid.New()
	m := mockbookinghandler.NewMockUseCase(t)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/cancel", nil)
	req = withChiParams(req, map[string]string{"bookingId": bookingID.String()})
	w := httptest.NewRecorder()
	h.CancelBooking(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCancelBooking_InvalidBookingID(t *testing.T) {
	userID := uuid.New()
	m := mockbookinghandler.NewMockUseCase(t)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodPost, "/bookings/not-a-uuid/cancel", nil)
	req = withUserID(req, userID.String())
	req = withChiParams(req, map[string]string{"bookingId": "not-a-uuid"})
	w := httptest.NewRecorder()
	h.CancelBooking(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCancelBooking_BookingNotFound(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().CancelBooking(mock.Anything, userID, bookingID).Return(nil, domain.ErrBookingNotFound)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/cancel", nil)
	req = withUserID(req, userID.String())
	req = withChiParams(req, map[string]string{"bookingId": bookingID.String()})
	w := httptest.NewRecorder()
	h.CancelBooking(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCancelBooking_Forbidden(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().CancelBooking(mock.Anything, userID, bookingID).Return(nil, domain.ErrForbidden)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/cancel", nil)
	req = withUserID(req, userID.String())
	req = withChiParams(req, map[string]string{"bookingId": bookingID.String()})
	w := httptest.NewRecorder()
	h.CancelBooking(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCancelBooking_InternalError(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	m := mockbookinghandler.NewMockUseCase(t)
	m.EXPECT().CancelBooking(mock.Anything, userID, bookingID).Return(nil, assert.AnError)

	h := bookinghandler.New(m)

	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/cancel", nil)
	req = withUserID(req, userID.String())
	req = withChiParams(req, map[string]string{"bookingId": bookingID.String()})
	w := httptest.NewRecorder()
	h.CancelBooking(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
