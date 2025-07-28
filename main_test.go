package main

import (
	"bytes"
	"fmt"
	db "k8s-backend/database"
	m "k8s-backend/model"
	svc "k8s-backend/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// go test -bench=.

func BenchmarkAPIHandler(b *testing.B) {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	b.ResetTimer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	for b.Loop() {
		router.ServeHTTP(res, req)
	}
}

func BenchmarkAPIData(b *testing.B) {
	router := gin.Default()
	router.GET("/api/data", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})
	// Perform the benchmark
	// Reset timer to exclude setup overhead
	b.ResetTimer()

	for b.Loop() {
		// Create a new HTTP request for the /api/data route
		req := httptest.NewRequest("GET", "/api/data", nil)
		resp := httptest.NewRecorder()

		// Serve the request using the Gin router
		router.ServeHTTP(resp, req)

		// Check the response status
		if resp.Code != http.StatusOK {
			b.Errorf("Expected status %v, got %v", http.StatusOK, resp.Code)
		}
	}
}

func BenchmarkCreateBook(b *testing.B) {
	bookSvc := &svc.BookService{
		DB: &db.Cache[m.Book]{},
		Cache: redis.NewClient(&redis.Options{
			Addr: "localhost:6379", // TODO: Config
		}),
	}
	bookSvc.Init()
	defer bookSvc.DB.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	bookSvc.SetupEndpoints(router)

	b.ResetTimer()

	for i := range b.N {
		book := fmt.Appendf(nil, `{"Id": %d, "Title": "E-Myth", "Author": "Michael Gerber", "Price": 15.99}`, i)

		req, err := http.NewRequestWithContext(
			b.Context(),
			http.MethodPost,
			"/api/v1/book",
			bytes.NewReader(book),
		)
		if err != nil {
			b.Error(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
	}
}
