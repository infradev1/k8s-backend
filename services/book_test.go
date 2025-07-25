package services

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	db "k8s-backend/database"
	"k8s-backend/model"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

//func TestMain(m *testing.M) {}

// TODO: table-driven tests
func TestGetBookHandler(t *testing.T) {
	bookSvc := &BookService{
		DB: &db.Cache[model.Book]{},
		Cache: redis.NewClient(&redis.Options{
			Addr: "localhost:6379", // TODO: Config
		}),
	}
	bookSvc.Init()
	defer bookSvc.DB.Close()

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		"/api/v1/books",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	bookSvc.SetupEndpoints(router)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, 200, rr.Code)
	t.Log(rr.Body.String())

	req, err = http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		"/api/v1/book/0",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	t.Log(rr.Body.String())

	req, err = http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		"/api/v1/book/10",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	t.Log(rr.Body.String())
	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCreateBookHandler(t *testing.T) {
	bookSvc := &BookService{
		DB: &db.Cache[model.Book]{},
		Cache: redis.NewClient(&redis.Options{
			Addr: "localhost:6379", // TODO: Config
		}),
	}
	bookSvc.Init()
	defer bookSvc.DB.Close()

	book := []byte(`{"Id": 15, "Title": "E-Myth", "Author": "Michael Gerber", "Price": 15.99}`)

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"/api/v1/book",
		bytes.NewReader(book),
	)
	if err != nil {
		t.Error(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	bookSvc.SetupEndpoints(router)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	t.Log(rr.Body.String())

	// t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestDeleteBookHandler(t *testing.T) {
	bookSvc := &BookService{
		DB: &db.Cache[model.Book]{},
		Cache: redis.NewClient(&redis.Options{
			Addr: "localhost:6379", // TODO: Config
		}),
	}
	bookSvc.Init()
	defer bookSvc.DB.Close()

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodDelete,
		"/api/v1/book?id=0",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	bookSvc.SetupEndpoints(router)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code)
	t.Log(rr.Body.String())

	req, err = http.NewRequestWithContext(
		t.Context(),
		http.MethodDelete,
		"/api/v1/book?id=10",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
	t.Log(rr.Body.String())
}
