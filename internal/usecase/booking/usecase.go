package booking

import (
	"context"
	"log/slog"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
)

type ConferenceService interface {
	CreateConferenceLink(ctx context.Context, bookingID uuid.UUID) (string, error)
}

type Repository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	Cancel(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
	ListByUserFuture(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error)
}

type SlotRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Slot, error)
}

type UseCase struct {
	bookingRepository Repository
	slotRepository    SlotRepository
	conferenceService ConferenceService
}

func New(bookingRepository Repository, slotRepository SlotRepository, conferenceService ConferenceService) *UseCase {
	return &UseCase{
		bookingRepository: bookingRepository,
		slotRepository:    slotRepository,
		conferenceService: conferenceService,
	}
}

func (uc *UseCase) CreateBooking(ctx context.Context, userID uuid.UUID, slotID uuid.UUID, createConferenceLink bool) (*domain.Booking, error) {
	slot, err := uc.slotRepository.GetByID(ctx, slotID)
	if err != nil {
		return nil, err
	}

	if slot.StartTime.Before(time.Now()) {
		return nil, domain.ErrPastSlot
	}

	booking := &domain.Booking{
		ID:     uuid.New(),
		SlotID: slotID,
		UserID: userID,
		Status: domain.BookingStatusActive,
	}

	if createConferenceLink && uc.conferenceService != nil {
		link, err := uc.conferenceService.CreateConferenceLink(ctx, booking.ID)
		if err != nil {
			slog.Warn("failed to create conference link", "bookingID", booking.ID, "error", err)
		} else {
			booking.ConferenceLink = &link
		}
	}

	if err := uc.bookingRepository.Create(ctx, booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (uc *UseCase) CancelBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*domain.Booking, error) {
	booking, err := uc.bookingRepository.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}

	if booking.Status == domain.BookingStatusCancelled {
		return booking, nil
	}

	if err := uc.bookingRepository.Cancel(ctx, bookingID); err != nil {
		return nil, err
	}

	booking.Status = domain.BookingStatusCancelled
	return booking, nil
}

func (uc *UseCase) GetMyBookings(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	bookings, err := uc.bookingRepository.ListByUserFuture(ctx, userID)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

func (uc *UseCase) ListBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	bookings, total, err := uc.bookingRepository.ListAll(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return bookings, total, nil
}
