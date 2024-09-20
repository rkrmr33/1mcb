package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/rkrmr33/1mcb/internal/assets"
	"github.com/rkrmr33/1mcb/pkg/util"
)

const OneMillion = 1_000

type (
	// Config is the server configuration
	Config struct {
		// Addr is the server bind address. Example: ":7888" for all interfaces on port 7888
		Addr string
		// MaxBodySize is the max size for the body of an incoming request in bytes
		MaxBodySize int64
	}

	// Server is the server interface
	Server interface {
		// Run starts the server
		Run() error
	}

	// server is the server implementation
	server struct {
		cfg          Config
		templates    *template.Template
		eventHandler *eventHandler
		state        state
	}
)

var _ Server = &server{}

// New creates a new server
func New(cfg Config) (Server, error) {
	fmt.Println("Preparing templates...")

	tpl, err := template.ParseFS(assets.Assets, "*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	stateBuf := make([]byte, OneMillion)
	// TODO load initial state from db

	return &server{
		cfg:          cfg,
		templates:    tpl,
		eventHandler: newEventHandler(),
		state:        newState(stateBuf),
	}, nil
}

// Run starts the server
func (s *server) Run() error {
	mux := s.prepareRoutes()

	fmt.Println("Starting server on", s.cfg.Addr)

	return http.ListenAndServe(s.cfg.Addr, util.LoggerMiddleware(mux))
}

func (s *server) prepareRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// index.html
	mux.HandleFunc("GET /{$}", s.rootHandler)

	// API
	mux.HandleFunc("GET /api/events", s.eventsHandler)
	mux.Handle("POST /api/toggle", http.MaxBytesHandler(http.HandlerFunc(s.toggleHandler), s.cfg.MaxBodySize))

	// Static assets handler
	mux.Handle("/static/", http.FileServer(http.FS(assets.Assets)))

	return mux
}
