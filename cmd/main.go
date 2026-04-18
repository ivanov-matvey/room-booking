// @title           Room Booking Service
// @version         1.0.0
// @description     Сервис бронирования переговорок
// @host            localhost:8080
// @BasePath        /
// @tag.name        Info
// @tag.name        Auth
// @tag.name        Rooms
// @tag.name        Schedules
// @tag.name        Slots
// @tag.name        Bookings
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/ivanov-matvey/room-booking/docs"
	"github.com/ivanov-matvey/room-booking/internal/conference"
	"github.com/ivanov-matvey/room-booking/internal/config"
	"github.com/ivanov-matvey/room-booking/internal/db"
	httpserver "github.com/ivanov-matvey/room-booking/internal/http"
	authhandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/auth"
	bookinghandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/booking"
	infohandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/info"
	roomhandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/room"
	schedulehandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/schedule"
	slothandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/slot"
	bookingrepo "github.com/ivanov-matvey/room-booking/internal/repository/booking"
	roomrepo "github.com/ivanov-matvey/room-booking/internal/repository/room"
	schedulerepo "github.com/ivanov-matvey/room-booking/internal/repository/schedule"
	slotrepo "github.com/ivanov-matvey/room-booking/internal/repository/slot"
	userrepo "github.com/ivanov-matvey/room-booking/internal/repository/user"
	"github.com/ivanov-matvey/room-booking/internal/seed"
	authusecase "github.com/ivanov-matvey/room-booking/internal/usecase/auth"
	bookingusecase "github.com/ivanov-matvey/room-booking/internal/usecase/booking"
	roomusecase "github.com/ivanov-matvey/room-booking/internal/usecase/room"
	scheduleusecase "github.com/ivanov-matvey/room-booking/internal/usecase/schedule"
	slotusecase "github.com/ivanov-matvey/room-booking/internal/usecase/slot"
)

func main() {
	doSeed := flag.Bool("seed", false, "seed the database with test data and exit")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL, db.PoolConfig{
		MaxConns:        cfg.DBMaxConns,
		MinConns:        cfg.DBMinConns,
		MaxConnLifetime: cfg.DBMaxConnLifetime,
		MaxConnIdleTime: cfg.DBMaxConnIdleTime,
	})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(pool); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	userRepo := userrepo.New(pool)
	roomRepo := roomrepo.New(pool)
	scheduleRepo := schedulerepo.New(pool)
	slotRepo := slotrepo.New(pool)
	bookingRepo := bookingrepo.New(pool)

	if err := userRepo.SeedDefaultUsers(ctx); err != nil {
		slog.Error("failed to seed default users", "error", err)
		os.Exit(1)
	}

	if *doSeed {
		if err := seed.Run(ctx, pool); err != nil {
			slog.Error("failed to seed database", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	confService := &conference.Stub{}

	authUC := authusecase.New(userRepo, cfg.JWTSecret, cfg.JWTExpiration)
	roomUC := roomusecase.New(roomRepo)
	scheduleUC := scheduleusecase.New(scheduleRepo, roomRepo)
	slotUC := slotusecase.New(slotRepo, roomRepo, scheduleRepo)
	bookingUC := bookingusecase.New(bookingRepo, slotRepo, confService)

	infoH := infohandler.New()
	authH := authhandler.New(authUC)
	roomH := roomhandler.New(roomUC)
	scheduleH := schedulehandler.New(scheduleUC)
	slotH := slothandler.New(slotUC)
	bookingH := bookinghandler.New(bookingUC)

	app := httpserver.New(
		cfg.JWTSecret,
		infoH,
		authH,
		roomH,
		scheduleH,
		slotH,
		bookingH,
	)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
		}
	}()

	slog.Info("starting server", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
