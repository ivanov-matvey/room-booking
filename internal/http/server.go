package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	swaggerFiles "github.com/swaggo/files/v2"
	"github.com/swaggo/swag"

	authhandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/auth"
	bookinghandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/booking"
	infohandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/info"
	roomhandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/room"
	schedulehandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/schedule"
	slothandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/slot"
	"github.com/ivanov-matvey/room-booking/internal/http/middleware"
)

type Server struct {
	router *chi.Mux
}

func New(
	jwtSecret string,
	infoH *infohandler.Handler,
	authH *authhandler.Handler,
	roomH *roomhandler.Handler,
	scheduleH *schedulehandler.Handler,
	slotH *slothandler.Handler,
	bookingH *bookinghandler.Handler,
) *Server {
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	authMiddleware := middleware.Auth(jwtSecret)

	r.Get("/_info", infoH.GetInfo)

	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		doc, _ := swag.ReadDoc()
		if _, err := w.Write([]byte(doc)); err != nil {
			slog.Error("failed to write swagger doc", "error", err)
		}
	})
	r.Get("/swagger/swagger-initializer.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		if _, err := w.Write([]byte(`window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/swagger/doc.json",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    plugins: [SwaggerUIBundle.plugins.DownloadUrl],
    layout: "StandaloneLayout"
  });
};`)); err != nil {
			slog.Error("failed to write swagger initializer", "error", err)
		}
	})
	fileServer := http.FileServer(http.FS(swaggerFiles.FS))
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/swagger", fileServer).ServeHTTP(w, r)
	})

	r.Post("/dummyLogin", authH.DummyLogin)
	r.Post("/register", authH.Register)
	r.Post("/login", authH.Login)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/rooms/list", roomH.ListRooms)
		r.With(middleware.RequireRole("admin")).Post("/rooms/create", roomH.CreateRoom)

		r.With(middleware.RequireRole("admin")).Post("/rooms/{roomId}/schedule/create", scheduleH.CreateSchedule)

		r.Get("/rooms/{roomId}/slots/list", slotH.ListSlots)

		r.With(middleware.RequireRole("user")).Post("/bookings/create", bookingH.CreateBooking)
		r.With(middleware.RequireRole("admin")).Get("/bookings/list", bookingH.ListBookings)
		r.With(middleware.RequireRole("user")).Get("/bookings/my", bookingH.GetMyBookings)
		r.With(middleware.RequireRole("user")).Post("/bookings/{bookingId}/cancel", bookingH.CancelBooking)
	})

	return &Server{router: r}
}

func (s *Server) Handler() http.Handler {
	return s.router
}
