package booking

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

func (r *Repository) Create(ctx context.Context, booking *domain.Booking) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO bookings (id, slot_id, user_id, status, conference_link)
		 VALUES ($1, $2, $3, $4, $5) RETURNING created_at`,
		booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink,
	).Scan(&booking.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrSlotBooked
		}
		return fmt.Errorf("create booking: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at FROM bookings WHERE id = $1`,
		id,
	)
	var b domain.Booking
	if err := row.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, fmt.Errorf("get booking by id: %w", err)
	}
	return &b, nil
}

func (r *Repository) Cancel(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE bookings SET status = $2 WHERE id = $1`,
		id, domain.BookingStatusCancelled,
	)
	if err != nil {
		return fmt.Errorf("cancel booking: %w", err)
	}
	return nil
}

func (r *Repository) ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	offset := (page - 1) * pageSize

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count bookings: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at FROM bookings
		 ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0, pageSize)
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list bookings: %w", err)
	}
	return bookings, total, nil
}

func (r *Repository) ListByUserFuture(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	rows, err := r.db.Query(ctx,
		`SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at
		 FROM bookings b
		 JOIN slots s ON s.id = b.slot_id
		 WHERE b.user_id = $1
		   AND s.start_time >= NOW()
		 ORDER BY s.start_time`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list user bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list user bookings: %w", err)
	}
	return bookings, nil
}

const pgUniqueViolation = "23505"

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
