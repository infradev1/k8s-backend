package services

import (
	"fmt"
	db "k8s-backend/database"
	m "k8s-backend/model"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type FleetService struct {
	DB db.Database[m.FleetHealthStatus]
}

func NewFleetService() *FleetService {
	return &FleetService{
		DB: &db.Cache[m.FleetHealthStatus]{
			Data: make(map[m.Region]*m.FleetHealthStatus),
		},
	}
}

func (f *FleetService) Init() {
	if err := f.DB.Initialize(); err != nil {
		slog.Error(err.Error())
		log.Fatal(fmt.Errorf("failed to initialize database: %w", err))
	}
}

func (f *FleetService) SetupEndpoints(r *gin.Engine) {
	// handlers can still be chained with a wrapper
	r.GET("/fleet", f.GetFleetHandler)
}

// Simulate fetching data from multiple sources (e.g., external APIs) and combine the results into a single response.
func (f *FleetService) GetFleetHandler(c *gin.Context) {
	net := make(chan bool)
	dc := make(chan bool)
	k8s := make(chan bool)

	go f.CheckNetworking(net)
	go f.CheckDataCenter(dc)
	go f.CheckKubernetes(k8s)

	c.JSON(http.StatusOK, &m.FleetHealthStatus{
		Networking: <-net,
		DataCenter: <-dc,
		Kubernetes: <-k8s,
	})
}

func (f *FleetService) CheckNetworking(ch chan bool) {
	time.Sleep(500 * time.Millisecond)
	ch <- true
}

func (f *FleetService) CheckDataCenter(ch chan bool) {
	time.Sleep(750 * time.Millisecond)
	ch <- true
}

func (f *FleetService) CheckKubernetes(ch chan bool) {
	time.Sleep(250 * time.Millisecond)
	ch <- false
}
