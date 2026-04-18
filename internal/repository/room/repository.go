package room

import (
	"context"
	"errors"
	"fmt"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	db DB
}

func New(db DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, room *domain.Room) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO rooms (id, name, description, capacity) VALUES ($1, $2, $3, $4) RETURNING created_at`,
		room.ID, room.Name, room.Description, room.Capacity,
	).Scan(&room.CreatedAt)
	if err != nil {
		return fmt.Errorf("create room: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context) ([]domain.Room, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	rooms := make([]domain.Room, 0)
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	return rooms, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`,
		id,
	)
	var room domain.Room
	if err := row.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, fmt.Errorf("get room by id: %w", err)
	}
	return &room, nil
}
