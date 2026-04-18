package seed

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ivanov-matvey/room-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var userID = uuid.MustParse("00000000-0000-0000-0000-000000000002")

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	slog.Info("seeding database...")

	roomIDs, err := seedRooms(ctx, pool)
	if err != nil {
		return fmt.Errorf("seed rooms: %w", err)
	}

	allSlots, err := seedSchedulesAndSlots(ctx, pool, roomIDs)
	if err != nil {
		return fmt.Errorf("seed schedules and slots: %w", err)
	}

	if err := seedBookings(ctx, pool, allSlots); err != nil {
		return fmt.Errorf("seed bookings: %w", err)
	}

	slog.Info("seeding done")
	return nil
}

func seedRooms(ctx context.Context, pool *pgxpool.Pool) ([]uuid.UUID, error) {
	rooms := []domain.Room{
		{ID: uuid.New(), Name: "Переговорка Красная", Capacity: intPtr(6)},
		{ID: uuid.New(), Name: "Переговорка Оранжевая", Capacity: intPtr(10)},
		{ID: uuid.New(), Name: "Переговорка Жёлтая", Capacity: intPtr(4)},
		{ID: uuid.New(), Name: "Переговорка Зелёная", Capacity: intPtr(12)},
		{ID: uuid.New(), Name: "Переговорка Голубая", Capacity: intPtr(8)},
		{ID: uuid.New(), Name: "Переговорка Синяя", Capacity: intPtr(6)},
		{ID: uuid.New(), Name: "Переговорка Фиолетовая", Capacity: intPtr(10)},
	}

	ids := make([]uuid.UUID, 0, len(rooms))
	for _, r := range rooms {
		_, err := pool.Exec(ctx,
			`INSERT INTO rooms (id, name, capacity) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			r.ID, r.Name, r.Capacity,
		)
		if err != nil {
			return nil, fmt.Errorf("insert room %s: %w", r.Name, err)
		}
		ids = append(ids, r.ID)
	}
	slog.Info("rooms seeded", "count", len(ids))
	return ids, nil
}

func seedSchedulesAndSlots(ctx context.Context, pool *pgxpool.Pool, roomIDs []uuid.UUID) ([]domain.Slot, error) {
	schedule := domain.Schedule{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	var allSlots []domain.Slot

	for _, roomID := range roomIDs {
		sch := schedule
		sch.ID = uuid.New()
		sch.RoomID = roomID

		_, err := pool.Exec(ctx,
			`INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
			 VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
			sch.ID, sch.RoomID, sch.DaysOfWeek, sch.StartTime, sch.EndTime,
		)
		if err != nil {
			return nil, fmt.Errorf("insert schedule for room %s: %w", roomID, err)
		}

		for i := -30; i <= 30; i++ {
			date := today.AddDate(0, 0, i)
			slots, err := domain.GenerateSlotsForDate(&sch, date)
			if err != nil {
				return nil, err
			}
			allSlots = append(allSlots, slots...)
		}
	}

	batch := make([]domain.Slot, 0, 100)
	flush := func() error {
		for _, s := range batch {
			_, err := pool.Exec(ctx,
				`INSERT INTO slots (id, room_id, start_time, end_time)
				 VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
				s.ID, s.RoomID, s.StartTime, s.EndTime,
			)
			if err != nil {
				return fmt.Errorf("insert slot: %w", err)
			}
		}
		batch = batch[:0]
		return nil
	}

	for _, s := range allSlots {
		batch = append(batch, s)
		if len(batch) >= 100 {
			if err := flush(); err != nil {
				return nil, err
			}
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}

	slog.Info("slots seeded", "count", len(allSlots))
	return allSlots, nil
}

func seedBookings(ctx context.Context, pool *pgxpool.Pool, slots []domain.Slot) error {
	now := time.Now().UTC()
	count := 0

	for i, s := range slots {
		if s.StartTime.After(now) {
			continue
		}
		if i%2 != 0 {
			continue
		}

		booking := domain.Booking{
			ID:     uuid.New(),
			SlotID: s.ID,
			UserID: userID,
			Status: domain.BookingStatusActive,
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO bookings (id, slot_id, user_id, status)
			 VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
			booking.ID, booking.SlotID, booking.UserID, booking.Status,
		)
		if err != nil {
			return fmt.Errorf("insert booking: %w", err)
		}
		count++
	}

	slog.Info("bookings seeded", "count", count)
	return nil
}

func intPtr(v int) *int { return &v }
