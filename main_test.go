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
