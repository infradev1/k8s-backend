package main

import (
	"context"
	"fmt"
	db "k8s-backend/database"
	m "k8s-backend/model"
	s "k8s-backend/server"
	svc "k8s-backend/services"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	bookSvc := &svc.BookService{
		DB: &db.Postgres[m.Book]{
			InitElements: []m.Book{
				{Title: "QM", Author: "Bohr", Price: 10.99},
				{Title: "QFT", Author: "Dirac", Price: 11.99},
				{Title: "GR", Author: "Einstein", Price: 12.99},
			},
		},
	}
	bookSvc.Init()
	defer bookSvc.DB.Close()

	go func() {
		server := &s.Server{
			Port: ":8081",
			Services: []s.Service{
				bookSvc,
			},
		}
		server.Run()
	}()

	<-ctx.Done()
	fmt.Println("exiting gracefully")
}
