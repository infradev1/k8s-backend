package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	db "k8s-backend/database"
	m "k8s-backend/model"
	s "k8s-backend/server"

	"github.com/stretchr/testify/require"
)

func TestGetBookHandler(t *testing.T) {
	bookSvc := &BookService{
		DB: &db.Cache[m.Book]{},
	}
	bookSvc.Init()
	defer bookSvc.DB.Close()

	// start server in separate goroutine
	go func() {
		server := &s.Server{
			Port:     ":8082",
			Services: []s.Service{bookSvc},
		}
		server.Run()
	}()
	time.Sleep(5 * time.Second)

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
}
