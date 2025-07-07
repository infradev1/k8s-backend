package main

import (
	db "k8s-backend/database"
	m "k8s-backend/model"
	s "k8s-backend/server"
)

func main() {
	db.InitDatabase()

	server := &s.Server{
		Port: ":8081",
		DB: &db.Cache[m.User]{
			Data: make(map[string]*m.User),
		},
	}
	server.Run()
}
