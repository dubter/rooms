package ports

import (
	"app-websocket/internal/config"
	"app-websocket/internal/ports/http/auth"
	"app-websocket/internal/ports/http/chat"
	"app-websocket/internal/ports/ws"
	"app-websocket/pkg/jwt"
	mwlogger "app-websocket/pkg/logger/middleware"
	"app-websocket/pkg/rate_limiter"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	logger          *slog.Logger
	server          *http.Server
	hub             *ws.Hub
	shutDownTimeout time.Duration
	certFilePath    string
	keyFilePath     string
}

func NewServer(config *config.HTTPConfig, authService auth.ServiceAuth, chatService chat.ServiceChatCache, chatPusher chat.ServiceChatPusher, roomsProvider chat.ServiceRoomsProvider, logger *slog.Logger, manager jwt.TokenManager, hub *ws.Hub) (*Server, error) {
	httpHandler := auth.NewHandler(logger, authService)
	wsHandler := chat.NewHandler(logger, chatService, chatPusher, roomsProvider)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      InitRouter(httpHandler, wsHandler, logger, &config.Limiter, manager),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &Server{
		hub:             hub,
		server:          server,
		shutDownTimeout: config.ShutdownTimeout,
		logger:          logger,
		certFilePath:    config.TLS.CertFilePath,
		keyFilePath:     config.TLS.KeyFilePath,
	}, nil
}

func InitRouter(auth *auth.Handler, chat *chat.Handler, logger *slog.Logger, limiter *config.Limiter, manager jwt.TokenManager) *chi.Mux {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300, // максимальный срок кэширования предварительных запросов
	}))

	mux.Use(rate_limiter.Limit(limiter.RPS, limiter.Burst, limiter.TTL, logger))
	mux.Use(middleware.Recoverer)
	mux.Use(mwlogger.Log(logger))

	mux.Post("/user/register", auth.Register)
	mux.Post("/user/login", auth.Login)
	mux.Post("/user/refresh", auth.RefreshTokens)

	mux.Route("/chat", func(r chi.Router) {
		r.Use(jwt.Validate(manager))

		r.Post("/rooms", chat.CreateRoom)
		r.Get("/rooms", chat.GetRooms)
		r.Get("/rooms/{id}/clients", chat.GetClients)
		r.HandleFunc("/rooms/{id}", chat.JoinRoom)
	})
	return mux
}

func (s *Server) Run(ctx context.Context) error {
	errResult := make(chan error)
	go func() {
		s.logger.Info(fmt.Sprintf("starting listening: %s", s.server.Addr))

		if s.certFilePath != "" && s.keyFilePath != "" {
			errResult <- s.server.ListenAndServeTLS(s.certFilePath, s.keyFilePath)
		} else {
			errResult <- s.server.ListenAndServe()
		}
	}()

	go func() {
		s.hub.Run(ctx)
	}()

	var err error
	select {
	case <-ctx.Done():
		return ctx.Err()

	case err = <-errResult:
	}
	return err
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutDownTimeout)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.logger.Error("failed to shutdown HTTP Server", slog.String("error", err.Error()))
	}
}
