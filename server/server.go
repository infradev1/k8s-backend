package server

import (
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	_ "k8s-backend/docs" // swag init | http://localhost:8081/swagger/index.html

	"github.com/gin-gonic/gin"
	f "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
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

	router.Use(loggingMiddleware, customHeaderMiddleware)

	rateLimiter := NewTokenBucket(5, 1*time.Second)
	router.Use(func(c *gin.Context) {
		if !rateLimiter.Allow() {
			c.String(http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	})

	// Set up Swagger UI to serve API documentation
	router.GET("/swagger/*any", gs.WrapHandler(f.Handler))

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

func loggingMiddleware(c *gin.Context) {
	start := time.Now()
	c.Next()
	latency := time.Since(start)
	log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
}

// customHeaderMiddleware adds a custom header to all responses
// Middleware in Gin is a function that takes a gin.Context and performs some operation
func customHeaderMiddleware(c *gin.Context) {
	// Add a custom header to the response
	// Headers set using c.Header will be included in the HTTP response
	c.Header("X-Custom-Header", "Middleware-Active")
	// Call the next middleware or the final handler in the chain
	c.Next()
}

type TokenBucket struct {
	Capacity   uint
	Tokens     uint
	Rate       time.Duration
	LastFilled time.Time
	sync.Mutex
}

func NewTokenBucket(capacity uint, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		Capacity:   capacity,
		Tokens:     capacity,
		Rate:       rate,
		LastFilled: time.Now().Local(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.Lock()
	defer tb.Unlock()

	elapsed := time.Since(tb.LastFilled)
	addTokens := uint(elapsed / tb.Rate) // refill if at least [1] second has elapsed
	tb.Tokens += addTokens
	if tb.Tokens > tb.Capacity {
		tb.Tokens = tb.Capacity
	}
	if addTokens > 0 {
		tb.LastFilled = time.Now()
	}

	if tb.Tokens > 0 {
		tb.Tokens--
		return true
	}

	return false
}
