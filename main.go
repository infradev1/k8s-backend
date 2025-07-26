package main

import (
	"context"
	"fmt"
	s "k8s-backend/server"
	svc "k8s-backend/services"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	go func() {
		slog.Info("starting pprof server on port 6060")
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	bookSvc := svc.NewBookService()
	bookSvc.Init()
	defer bookSvc.DB.Close()

	go func() {
		s.NewServer(":8081", []s.Service{bookSvc}).Run()
	}()

	<-ctx.Done()
	slog.Info("exiting gracefully")
}
