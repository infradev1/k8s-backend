package services

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	db "k8s-backend/database"
	"k8s-backend/model"
	s "k8s-backend/server"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	bookSvc := &BookService{
		DB: &db.Cache[model.Book]{},
	}
	bookSvc.Init()

	// start server in separate goroutine
	go func() {
		server := &s.Server{
			Port:     ":8082",
			Services: []s.Service{bookSvc},
		}
		server.Run()
	}()
	time.Sleep(5 * time.Second)

	exitCode := m.Run()

	bookSvc.DB.Close()

	os.Exit(exitCode)
}

func TestGetBookHandler(t *testing.T) {
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"http://localhost:8082/books?id=0",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	require.Equal(t, 200, rsp.StatusCode)

	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"http://localhost:8082/books?id=10",
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	rsp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	require.Equal(t, 404, rsp.StatusCode)
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
		"http://localhost:8082/books",
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
