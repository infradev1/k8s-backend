package main

import (
	"context"
	"fmt"
	db "k8s-backend/database"
	m "k8s-backend/model"
	s "k8s-backend/server"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	go func() {
		server := &s.BookServer{
			Port: ":8081",
			DB: &db.Postgres[m.Book]{
				InitElements: []m.Book{
					{Title: "QM", Author: "Bohr", Price: 10.99},
					{Title: "QFT", Author: "Dirac", Price: 11.99},
					{Title: "GR", Author: "Einstein", Price: 12.99},
				},
			},
		}
		server.Run()
	}()

	<-ctx.Done()
	fmt.Println("exiting gracefully")
}
