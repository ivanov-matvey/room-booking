package room

import (
	"context"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, room *domain.Room) error
	List(ctx context.Context) ([]domain.Room, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error)
}

type UseCase struct {
	roomRepository Repository
}

func New(roomRepository Repository) *UseCase {
	return &UseCase{roomRepository: roomRepository}
}

func (uc *UseCase) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (*domain.Room, error) {
	room := &domain.Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
	}

	if err := uc.roomRepository.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

func (uc *UseCase) ListRooms(ctx context.Context) ([]domain.Room, error) {
	return uc.roomRepository.List(ctx)
}
