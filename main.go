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
		s.NewServer(":8081", []s.Service{bookSvc}).Run()
	}()

	<-ctx.Done()
	slog.Info("exiting gracefully")
}
