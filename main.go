package main

import (
	"context"
	s "k8s-backend/server"
	svc "k8s-backend/services"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	bookSvc := svc.NewBookService()
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
	slog.Info("exiting gracefully")
}
