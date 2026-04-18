package httpserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"
	httpserver "github.com/ivanov-matvey/room-booking/internal/http"
	authhandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/auth"
	bookinghandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/booking"
	infohandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/info"
	roomhandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/room"
	schedulehandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/schedule"
	slothandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/slot"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type stubAuthUC struct{}

func (s *stubAuthUC) DummyLogin(_ context.Context, _ string) (string, error) { return "", nil }
func (s *stubAuthUC) Register(_ context.Context, _, _, _ string) (*domain.User, error) {
	return nil, nil
}
func (s *stubAuthUC) Login(_ context.Context, _, _ string) (string, error) { return "", nil }

type stubRoomUC struct{}

func (s *stubRoomUC) CreateRoom(_ context.Context, _ string, _ *string, _ *int) (*domain.Room, error) {
	return nil, nil
}
func (s *stubRoomUC) ListRooms(_ context.Context) ([]domain.Room, error) { return nil, nil }

type stubScheduleUC struct{}

func (s *stubScheduleUC) CreateSchedule(_ context.Context, _ uuid.UUID, _ []int, _, _ string) (*domain.Schedule, error) {
	return nil, nil
}

type stubSlotUC struct{}

func (s *stubSlotUC) GetAvailableSlots(_ context.Context, _ uuid.UUID, _ time.Time) ([]domain.Slot, error) {
	return nil, nil
}

type stubBookingUC struct{}

func (s *stubBookingUC) CreateBooking(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ bool) (*domain.Booking, error) {
	return nil, nil
}
func (s *stubBookingUC) CancelBooking(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*domain.Booking, error) {
	return nil, nil
}
func (s *stubBookingUC) GetMyBookings(_ context.Context, _ uuid.UUID) ([]domain.Booking, error) {
	return nil, nil
}
func (s *stubBookingUC) ListBookings(_ context.Context, _, _ int) ([]domain.Booking, int, error) {
	return nil, 0, nil
}

func newTestServer(t *testing.T) *httpserver.Server {
	t.Helper()
	return httpserver.New(
		"test-secret",
		infohandler.New(),
		authhandler.New(&stubAuthUC{}),
		roomhandler.New(&stubRoomUC{}),
		schedulehandler.New(&stubScheduleUC{}),
		slothandler.New(&stubSlotUC{}),
		bookinghandler.New(&stubBookingUC{}),
	)
}

func TestServer_New(t *testing.T) {
	server := newTestServer(t)
	assert.NotNil(t, server)
}

func TestServer_Handler(t *testing.T) {
	server := newTestServer(t)
	handler := server.Handler()
	assert.NotNil(t, handler)
}

func TestServer_InfoRoute(t *testing.T) {
	server := newTestServer(t)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/_info", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_UnknownRoute(t *testing.T) {
	server := newTestServer(t)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
