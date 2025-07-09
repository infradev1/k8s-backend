package services

import (
	"encoding/json"
	"fmt"
	db "k8s-backend/database"
	m "k8s-backend/model"
	"log"
	"log/slog"
	"net/http"
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
	http.HandleFunc("/books", s.GetBookHandler)
}

func (s *BookService) GetBookHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Query parameter 'id' must be provided", http.StatusBadRequest)
		return
	}

	book, err := s.DB.Get(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("id %s not found: %v", id, err), http.StatusNotFound)
		return
	}

	data, err := json.MarshalIndent(book, "", "  ")
	if err != nil {
		http.Error(w, "Error marshaling struct into JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		http.Error(w, "Error writing response JSON", http.StatusInternalServerError)
		return
	}
}
