package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	CreateWithSlots(ctx context.Context, schedule *domain.Schedule, slots []domain.Slot) error
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error)
}

type RoomRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error)
}

type UseCase struct {
	scheduleRepository Repository
	roomRepository     RoomRepository
}

const slotGenerationDays = 7

func New(scheduleRepository Repository, roomRepository RoomRepository) *UseCase {
	return &UseCase{
		scheduleRepository: scheduleRepository,
		roomRepository:     roomRepository,
	}
}

func (uc *UseCase) CreateSchedule(ctx context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*domain.Schedule, error) {
	if _, err := uc.roomRepository.GetByID(ctx, roomID); err != nil {
		return nil, err
	}

	existing, err := uc.scheduleRepository.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrScheduleExists
	}

	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	now := time.Now().UTC()
	allSlots := make([]domain.Slot, 0)
	for i := 0; i < slotGenerationDays; i++ {
		date := now.AddDate(0, 0, i)
		daySlots, err := domain.GenerateSlotsForDate(schedule, date)
		if err != nil {
			return nil, fmt.Errorf("generate slots: %w", err)
		}
		allSlots = append(allSlots, daySlots...)
	}

	if err := uc.scheduleRepository.CreateWithSlots(ctx, schedule, allSlots); err != nil {
		return nil, err
	}

	return schedule, nil
}

func (uc *UseCase) GetScheduleByRoom(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error) {
	return uc.scheduleRepository.GetByRoomID(ctx, roomID)
}
