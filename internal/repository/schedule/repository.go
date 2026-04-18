package schedule

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
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	db DB
}

func New(db DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, schedule *domain.Schedule) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES ($1, $2, $3, $4, $5)`,
		schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrScheduleExists
		}
		return fmt.Errorf("create schedule: %w", err)
	}
	return nil
}

func (r *Repository) CreateWithSlots(ctx context.Context, schedule *domain.Schedule, slots []domain.Slot) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := tx.QueryRow(ctx,
		`INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES ($1, $2, $3, $4, $5) RETURNING created_at`,
		schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime,
	).Scan(&schedule.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrScheduleExists
		}
		return fmt.Errorf("create schedule: %w", err)
	}

	if len(slots) > 0 {
		batch := &pgx.Batch{}
		for _, slot := range slots {
			batch.Queue(
				`INSERT INTO slots (id, room_id, start_time, end_time) VALUES ($1, $2, $3, $4) ON CONFLICT (room_id, start_time) DO NOTHING`,
				slot.ID, slot.RoomID, slot.StartTime, slot.EndTime,
			)
		}
		br := tx.SendBatch(ctx, batch)
		for range slots {
			if _, err := br.Exec(); err != nil {
				_ = br.Close()
				return fmt.Errorf("insert slot: %w", err)
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("close batch: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (r *Repository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, room_id, days_of_week, start_time, end_time, created_at FROM schedules WHERE room_id = $1`,
		roomID,
	)
	var sch domain.Schedule
	if err := row.Scan(&sch.ID, &sch.RoomID, &sch.DaysOfWeek, &sch.StartTime, &sch.EndTime, &sch.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get schedule by room id: %w", err)
	}
	return &sch, nil
}

const pgUniqueViolation = "23505"

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
