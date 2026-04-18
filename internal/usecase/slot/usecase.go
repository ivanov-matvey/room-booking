package slot

import (
	"context"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	ListAvailable(ctx context.Context, roomID uuid.UUID, from, to time.Time) ([]domain.Slot, error)
	BulkUpsert(ctx context.Context, slots []domain.Slot) error
	CountByRoomAndDate(ctx context.Context, roomID uuid.UUID, from, to time.Time) (int, error)
}

type RoomRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error)
}

type ScheduleRepository interface {
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error)
}

type UseCase struct {
	slotRepository     Repository
	roomRepository     RoomRepository
	scheduleRepository ScheduleRepository
}

const (
	maxLazyGenerationDays = 90
	pastDaysAllowed       = 1
)

func New(slotRepository Repository, roomRepository RoomRepository, scheduleRepository ScheduleRepository) *UseCase {
	return &UseCase{
		slotRepository:     slotRepository,
		roomRepository:     roomRepository,
		scheduleRepository: scheduleRepository,
	}
}

func (uc *UseCase) GetAvailableSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	if _, err := uc.roomRepository.GetByID(ctx, roomID); err != nil {
		return nil, err
	}

	schedule, err := uc.scheduleRepository.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return []domain.Slot{}, nil
	}

	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)

	daysDiff := int(time.Until(from).Hours() / 24)
	if daysDiff >= -pastDaysAllowed && daysDiff <= maxLazyGenerationDays {
		count, err := uc.slotRepository.CountByRoomAndDate(ctx, roomID, from, to)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			slots, err := domain.GenerateSlotsForDate(schedule, date)
			if err != nil {
				return nil, err
			}
			if len(slots) > 0 {
				if err := uc.slotRepository.BulkUpsert(ctx, slots); err != nil {
					return nil, err
				}
			}
		}
	}

	slots, err := uc.slotRepository.ListAvailable(ctx, roomID, from, to)
	if err != nil {
		return nil, err
	}

	return slots, nil
}
