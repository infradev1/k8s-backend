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
	http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.GetBookHandler(w, r)
		} else {
			http.Error(w, fmt.Sprintf("%s not recognized; use GET", r.Method), http.StatusMethodNotAllowed)
			//fmt.Fprintf(w, "%s not recognized; use GET", r.Method)
		}
	})

	http.HandleFunc("/book", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.CreateBookHandler(w, r)
		case http.MethodDelete:
			s.DeleteBookHandler(w, r)
		case http.MethodPatch:
			s.UpdateBookHandler(w, r)
		default:
			http.Error(w, fmt.Sprintf("%s not recognized; use POST, PATCH, or DELETE", r.Method), http.StatusMethodNotAllowed)
			//fmt.Fprintf(w, "%s not recognized; use POST or DELETE", r.Method)
		}
	})
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
	if _, err := fmt.Fprintf(w, "%s created successfully with ID %d", book.Title, book.Id); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

func (s *BookService) UpdateBookHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "query parameter 'id' must be provided", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	// Decode the JSON body into a map of fields to update
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := s.DB.Update(id, updates); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update book: %v", err), http.StatusInternalServerError)
		return
	}

	// Return 204 No Content to indicate a successful update
	w.WriteHeader(http.StatusNoContent)
}

func (s *BookService) DeleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "query parameter 'id' must be provided", http.StatusBadRequest)
		return
	}

	if err := s.DB.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Book ID %s deleted", id)
}

func ValidateBook(book *m.Book) error {
	return nil
}
