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
	"time"

	"github.com/gin-gonic/gin"
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

func (s *BookService) SetupEndpoints(r *gin.Engine) {
	// handlers can still be chained with a wrapper
	r.GET("/books", s.GetBooksHandler)
	r.GET("/book/:id", s.GetBookHandler)
	r.POST("/book", s.CreateBookHandler)
	r.PATCH("/book", s.UpdateBookHandler)
	r.DELETE("/book", s.DeleteBookHandler)
}

func (s *BookService) GetBooksHandler(c *gin.Context) {
	books, err := s.DB.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, books)
}

func (s *BookService) GetBookHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		http.Error(c.Writer, "id path parameter must be provided", http.StatusBadRequest)
		return
	}

	book, err := s.DB.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"id": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

func (s *BookService) CreateBookHandler(c *gin.Context) {
	var book m.Book
	if err := c.ShouldBindBodyWithJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book.CreatedAt = time.Now().Format(time.RFC3339)

	if err := ValidateBook(&book); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.DB.Insert("", &book); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.String(http.StatusCreated, "%s created successfully with ID %d", book.Title, book.Id)
}

func (s *BookService) UpdateBookHandler(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		http.Error(c.Writer, "query parameter 'id' must be provided", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	// Decode the JSON body into a map of fields to update
	if err := json.NewDecoder(c.Request.Body).Decode(&updates); err != nil {
		http.Error(c.Writer, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := s.DB.Update(id, updates); err != nil {
		http.Error(c.Writer, fmt.Sprintf("Failed to update book: %v", err), http.StatusInternalServerError)
		return
	}

	// Return 204 No Content to indicate a successful update
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (s *BookService) DeleteBookHandler(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		http.Error(c.Writer, "query parameter 'id' must be provided", http.StatusBadRequest)
		return
	}

	// TODO: Get first to see if it exists

	if err := s.DB.Delete(id); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

func ValidateBook(book *m.Book) error {
	return nil
}
