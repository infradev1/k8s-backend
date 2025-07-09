package server

import (
	"log"
	"log/slog"
	"net/http"
)

type Service interface {
	Init()
	SetupEndpoints()
}

type Server struct {
	Port     string
	Services []Service
}

func (s *Server) Run() {
	for _, svc := range s.Services {
		slog.Info("setting up endpoints")
		svc.SetupEndpoints()
	}

	slog.Info("starting server...")
	if err := http.ListenAndServe(s.Port, nil); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
