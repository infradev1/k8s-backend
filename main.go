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

	db.InitDatabase()
	defer db.CloseDatabase()

	go func() {
		server := &s.Server{
			Port: ":8081",
			DB: &db.Cache[m.User]{
				Data: make(map[string]*m.User),
			},
		}
		server.Run()
	}()

	<-ctx.Done()
	fmt.Println("exiting gracefully")
}
