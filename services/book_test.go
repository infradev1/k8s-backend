package services

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	db "k8s-backend/database"
	"k8s-backend/model"

	"github.com/stretchr/testify/require"
)

//func TestMain(m *testing.M) {}

// TODO: table-driven tests
func TestGetBookHandler(t *testing.T) {
	bookSvc := &BookService{
		DB: &db.Cache[model.Book]{},
	}
	bookSvc.Init()
	defer bookSvc.DB.Close()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"http://localhost:8082/books?id=0",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bookSvc.GetBookHandler)
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	t.Log(rr.Body.String())

	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"http://localhost:8082/books?id=10",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(bookSvc.GetBookHandler)
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)

	t.Log(rr.Body.String())

	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"http://localhost:8082/books",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(bookSvc.GetBookHandler)
	handler.ServeHTTP(rr, req)
	require.Equal(t, 200, rr.Code)

	t.Log(rr.Body.String())
}

func TestCreateBookHandler(t *testing.T) {
	bookSvc := &BookService{
		DB: &db.Cache[model.Book]{},
	}
	bookSvc.Init()
	defer bookSvc.DB.Close()

	book := []byte(`{"Id": 15, "Title": "E-Myth", "Author": "Michael Gerber", "Price": 15.99}`)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"http://localhost:8082/book",
		bytes.NewReader(book),
	)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bookSvc.CreateBookHandler)
	handler.ServeHTTP(rr, req)

	t.Log(rr.Body.String())

	// t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestDeleteBookHandler(t *testing.T) {
	bookSvc := &BookService{
		DB: &db.Cache[model.Book]{},
	}
	bookSvc.Init()
	defer bookSvc.DB.Close()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		"http://localhost:8082/books?id=0",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bookSvc.DeleteBookHandler)
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	t.Log(rr.Body.String())

	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		"http://localhost:8082/books?id=10",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(bookSvc.DeleteBookHandler)
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)

	t.Log(rr.Body.String())
}
