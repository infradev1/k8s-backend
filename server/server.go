package server

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Service interface {
	Init()
	SetupEndpoints(r *gin.Engine)
}

type Server struct {
	Router   *gin.Engine
	Port     string
	Services []Service
}

func NewServer(port string, services []Service) *Server {
	router := gin.Default()

	router.Use(LoggingMiddleware)

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Gin server healthy")
	})
	router.GET("/svc", func(c *gin.Context) {
		// get query parameter
		name := c.DefaultQuery("name", "Book")
		c.JSON(http.StatusOK, gin.H{"service": name})
	})

	slog.Info("Gin router", "base path: %s", router.BasePath())

	return &Server{
		Router:   router,
		Port:     port,
		Services: services,
	}
}

func (s *Server) Run() {
	for _, svc := range s.Services {
		slog.Info("setting up endpoints")
		svc.SetupEndpoints(s.Router)
	}

	slog.Info("starting server", "port", s.Port)
	if err := s.Router.Run(s.Port); err != nil {
		log.Fatal(err)
	}
	//if err := http.ListenAndServe(s.Port, nil); err != http.ErrServerClosed {
	//	log.Fatal(err)
	//}
}

func LoggingMiddleware(c *gin.Context) {
	start := time.Now()
	c.Next()
	latency := time.Since(start)
	log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
}
