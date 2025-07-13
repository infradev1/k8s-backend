package services

import (
	"encoding/json"
	"fmt"
	db "k8s-backend/database"
	m "k8s-backend/model"
	"log"
	"log/slog"
	"net/http"
	"strings"
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
	http.HandleFunc("/book", s.CreateBookHandler)
	http.HandleFunc("/books/{id}", s.DeleteBookHandler)
}

func (s *BookService) GetBookHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	w.Header().Add("Content-Type", "application/json")
	var data []byte

	if id == "" {
		books, err := s.DB.GetAll()
		if err != nil {
			http.Error(w, fmt.Sprintf("Query parameter 'id' must be provided for single book, otherwise: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(books)
		return
	}

	book, err := s.DB.Get(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("id %s not found: %v", id, err), http.StatusNotFound)
		return
	}

	data, err = json.MarshalIndent(book, "", "  ")
	if err != nil {
		http.Error(w, "Error marshaling struct into JSON", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(data); err != nil {
		http.Error(w, "Error writing response JSON", http.StatusInternalServerError)
		return
	}
}

func (s *BookService) CreateBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusBadRequest)
		return
	}

	var book m.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Request body must contain title, author, and price", http.StatusBadRequest)
		return
	}

	if err := ValidateBook(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.DB.Insert("", &book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err := fmt.Fprintf(w, "%s created successfully", book.Title); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

func (s *BookService) DeleteBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE required to delete a book by ID", http.StatusBadRequest)
		return
	}

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if err := s.DB.Delete(segments[2]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Book ID %s deleted", segments[2])
}

func ValidateBook(book *m.Book) error {
	return nil
}
