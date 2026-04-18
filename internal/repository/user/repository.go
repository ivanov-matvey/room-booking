package user

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
}

type Repository struct {
	db DB
}

func New(db DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user *domain.User) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (id, email, role, password_hash) VALUES ($1, $2, $3, $4) RETURNING created_at`,
		user.ID, user.Email, user.Role, user.PasswordHash,
	).Scan(&user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

const pgUniqueViolation = "23505"

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, email, role, password_hash, created_at FROM users WHERE email = $1`,
		email,
	)
	var u domain.User
	if err := row.Scan(&u.ID, &u.Email, &u.Role, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *Repository) SeedDefaultUsers(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, role) VALUES
		    ('00000000-0000-0000-0000-000000000001', 'admin@example.com', 'admin'),
		    ('00000000-0000-0000-0000-000000000002', 'user@example.com', 'user')
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed default users: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, email, role, password_hash, created_at FROM users WHERE id = $1`,
		id,
	)
	var u domain.User
	if err := row.Scan(&u.ID, &u.Email, &u.Role, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}
