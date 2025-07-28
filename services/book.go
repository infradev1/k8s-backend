package services

import (
	"encoding/json"
	"fmt"
	db "k8s-backend/database"
	m "k8s-backend/model"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type BookService struct {
	DB    db.Database[m.Book]
	Cache *redis.Client
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
		Cache: redis.NewClient(&redis.Options{
			Addr: "localhost:6379", // TODO: Config
		}),
	}
}

func (s *BookService) Init() {
	if err := s.DB.Initialize(); err != nil {
		slog.Error(err.Error())
		log.Fatal(fmt.Errorf("failed to initialize database: %w", err))
	}
}

func (s *BookService) SetupEndpoints(r *gin.Engine) {
	v1 := r.Group("api/v1")
	{
		// handlers can still be chained with a wrapper
		v1.GET("/books", s.GetBooksHandler)
		v1.GET("/book/:id", s.GetBookHandler)
		v1.POST("/book", s.CreateBookHandler)
		v1.PATCH("/book", s.UpdateBookHandler)
		v1.DELETE("/book", s.DeleteBookHandler)
	}

	// Versioning ensures backward compatibility by segregating changes into distinct API versions.
	// Overall, this structure supports multiple versions of the same endpoint, enabling incremental updates without breaking existing clients.
	//v2 := router.Group("/api/v2")
	//{
	//	v2.GET("/notes", getEnhancedBookHandler)
	//}
}

// GetBooksHandler godoc
// @Summary Get all books
// @Description Retrieve a list of all available books
// @Tags books
// @Produce json
// @Success 200 {array} m.Book
// @Router /api/v1/books [get]
func (s *BookService) GetBooksHandler(c *gin.Context) {
	// extract query parameters
	l := c.DefaultQuery("limit", "10")
	o := c.DefaultQuery("offset", "0")

	filters := new(m.Filters[m.Book])
	filters.Model = new(m.Book)
	filters.Model.Title = c.Query("title")
	filters.Model.Author = c.Query("author")
	if price := c.Query("price"); price != "" {
		n, err := strconv.ParseFloat(price, 32)
		if err != nil {
			c.String(http.StatusBadRequest, "'price' query parameter must be a number >= 0")
			return
		}
		filters.Model.Price = n
	}

	filters.SortBy = c.DefaultQuery("sortBy", "title")
	order := c.DefaultQuery("order", "ASC")
	if order != "ASC" && order != "DESC" {
		c.String(http.StatusBadRequest, "'order' query parameter must be ASC or DESC")
		return
	}
	filters.Order = order

	limit, err := strconv.Atoi(l)
	if err != nil || limit <= 0 {
		c.String(http.StatusBadRequest, "response limit must be a number greater than zero: %v", err)
		return
	}
	filters.Limit = limit

	offset, err := strconv.Atoi(o)
	if err != nil || offset < 0 {
		c.String(http.StatusBadRequest, fmt.Sprintf("response offset must be a non-negative number: %v", err))
		return
	}
	filters.Offset = offset

	// each request gets its own unbuffered channel
	queue := make(chan *m.Result)

	go func() {
		books, err := s.DB.GetAll(filters)
		queue <- &m.Result{Value: books, Error: err}
	}()

	r := <-queue
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, r.Error.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":     r.Value,
		"metadata": filters,
	})
}

func (s *BookService) GetBookHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		http.Error(c.Writer, "id path parameter must be provided", http.StatusBadRequest)
		return
	}

	// check Redis cache first
	key := fmt.Sprintf("book:%s", id)
	val, err := s.Cache.Get(c, key).Result()
	if err == nil {
		// cache hit
		var book m.Book
		if err := json.Unmarshal([]byte(val), &book); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, book)
		return
	} else if err != redis.Nil {
		// any error other than a cache miss
		c.String(http.StatusInternalServerError, "redis get error: %w", err)
		return
	}

	// cache miss
	book, err := s.DB.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	data, err := json.Marshal(book)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.Cache.Set(c, key, data, 24*time.Hour).Err(); err != nil {
		c.String(http.StatusInternalServerError, "redis set error: %w", err)
		return
	}

	c.JSON(http.StatusOK, book)
}

// CreateBookHandler godoc
// @Summary Create a new book
// @Description Add a new book entry
// @Tags books
// @Accept json
// @Produce json
// @Param book body model.Book true "Book data"
// @Success 201 {object} model.Book
// @Failure 400 {object} error
// @Router /api/v1/book [post]
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
