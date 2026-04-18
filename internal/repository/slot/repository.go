package slot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	db DB
}

func New(db DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BulkUpsert(ctx context.Context, slots []domain.Slot) error {
	if len(slots) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	batch := &pgx.Batch{}
	for _, slot := range slots {
		batch.Queue(
			`INSERT INTO slots (id, room_id, start_time, end_time)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (room_id, start_time) DO NOTHING`,
			slot.ID, slot.RoomID, slot.StartTime, slot.EndTime,
		)
	}

	br := tx.SendBatch(ctx, batch)
	for range slots {
		if _, err := br.Exec(); err != nil {
			_ = br.Close()
			return fmt.Errorf("upsert slot: %w", err)
		}
	}
	if err := br.Close(); err != nil {
		return fmt.Errorf("close batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (r *Repository) ListAvailable(ctx context.Context, roomID uuid.UUID, from, to time.Time) ([]domain.Slot, error) {
	rows, err := r.db.Query(ctx,
		`SELECT s.id, s.room_id, s.start_time, s.end_time, s.created_at
		 FROM slots s
		 WHERE s.room_id = $1
		   AND s.start_time >= $2
		   AND s.start_time < $3
		   AND NOT EXISTS (
		     SELECT 1 FROM bookings b WHERE b.slot_id = s.id AND b.status = $4
		   )
		 ORDER BY s.start_time`,
		roomID, from, to, domain.BookingStatusActive,
	)
	if err != nil {
		return nil, fmt.Errorf("list available slots: %w", err)
	}
	defer rows.Close()

	slots := make([]domain.Slot, 0)
	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.StartTime, &slot.EndTime, &slot.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan slot: %w", err)
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list available slots: %w", err)
	}
	return slots, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Slot, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, room_id, start_time, end_time, created_at FROM slots WHERE id = $1`,
		id,
	)
	var slot domain.Slot
	if err := row.Scan(&slot.ID, &slot.RoomID, &slot.StartTime, &slot.EndTime, &slot.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSlotNotFound
		}
		return nil, fmt.Errorf("get slot by id: %w", err)
	}
	return &slot, nil
}

func (r *Repository) CountByRoomAndDate(ctx context.Context, roomID uuid.UUID, from, to time.Time) (int, error) {
	var count int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM slots WHERE room_id = $1 AND start_time >= $2 AND start_time < $3`,
		roomID, from, to,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("count slots by room and date: %w", err)
	}
	return count, nil
}
