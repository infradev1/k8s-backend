package services

import (
	"fmt"
	db "k8s-backend/database"
	m "k8s-backend/model"
	"log"
	"log/slog"
)

type BookService struct {
	DB db.Database[m.Book]
}

func NewBookService() *BookService {
	return &BookService{
		DB: &db.Postgres[m.Book]{
			InitElements: []m.Book{
				{Title: "QM", Author: "Bohr", Price: 10.99},
				{Title: "QFT", Author: "Dirac", Price: 11.99},
				{Title: "GR", Author: "Einstein", Price: 12.99},
			},
		},
	}
}

func (s *BookService) Init() {
	if err := s.DB.Initialize(); err != nil {
		slog.Error(err.Error())
		log.Fatal(fmt.Errorf("failed to initialize database: %w", err))
	}
}

func (s *BookService) SetupEndpoints() {
	// TODO
}
