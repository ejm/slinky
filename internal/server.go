package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	sloghttp "github.com/samber/slog-http"
)

type Server struct {
	ctx        context.Context
	httpClient *http.Client
	config     Config
	db         *sql.DB
	logger     *slog.Logger

	handler http.Handler
}

// Run the HTTP server
func (s *Server) Run() error {
	return http.ListenAndServe(s.config.ListenAddr, s.handler)
}

// Create a new Server from a config
func NewServer(config Config) (*Server, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logConfig := sloghttp.Config{
		WithSpanID:    true,
		WithTraceID:   true,
		WithRequestID: true,
	}

	db, err := sql.Open("postgres", config.DatabaseUri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	httpClient := &http.Client{
		Timeout: time.Second * 1,
	}

	router := http.NewServeMux()
	handler := sloghttp.Recovery(router)
	handler = sloghttp.NewWithConfig(logger, logConfig)(handler)

	server := &Server{
		ctx:        context.Background(),
		httpClient: httpClient,
		config:     config,
		db:         db,
		logger:     logger,
		handler:    handler,
	}

	router.Handle("/api/links", server.authMiddleware(http.HandlerFunc(server.createShortLinkHandler)))
	router.HandleFunc("/{link}", server.shortLinkRedirectHandler)

	return server, nil
}
