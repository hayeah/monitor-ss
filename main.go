package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
)

type HTTPPingStatus struct {
	Code     int
	Status   string
	Duration time.Duration
	Err      error
}

type serverConfigs map[string]SSServerConfig

type Monitor struct {
	ctx           context.Context
	portBase      int
	serverConfigs serverConfigs

	statuses map[uint]*HTTPPingStatus
}

func NewMonitor(ctx context.Context, portBase int, servers serverConfigs) *Monitor {
	return &Monitor{
		ctx:           ctx,
		portBase:      portBase,
		serverConfigs: servers,
		statuses:      make(map[uint]*HTTPPingStatus),
	}
}

func (m *Monitor) StartServers() {
	var i = 0

	ctx := context.Background()

	for name, config := range m.serverConfigs {

		localPort := m.portBase + i
		config := config
		id := fmt.Sprintf("%s@%d", name, localPort)

		go func() {
			s := &ssService{
				id:        id,
				localPort: localPort,
				config:    config,
			}

			s.start(ctx)
		}()

		i++
	}

	<-m.ctx.Done()
}

func checkURL(client *http.Client, url string) (pingResult HTTPPingStatus) {
	// http.DefaultTransport
	var err error

	startAt := time.Now()
	defer func() {
		endAt := time.Now()
		pingResult.Duration = endAt.Sub(startAt)
	}()

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		pingResult.Err = err
		return
	}

	// client := http.DefaultClient

	res, err := client.Do(req)
	if err != nil {
		pingResult.Err = err
		return
	}

	pingResult.Status = res.Status
	pingResult.Code = res.StatusCode

	return
}

func startMonitor(configFile string) (err error) {
	f, err := os.Open(configFile)
	if err != nil {
		return
	}
	defer f.Close()

	var serverConfigs serverConfigs
	dec := json.NewDecoder(f)
	err = dec.Decode(&serverConfigs)
	if err != nil {
		return
	}

	ctx := context.Background()

	mon := NewMonitor(ctx, 1080, serverConfigs)
	mon.StartServers()

	return nil
}

func main() {

	configFile := os.Args[1]

	go metrics.Log(
		metrics.DefaultRegistry,
		10*time.Second,
		log.New(os.Stderr, "metrics: ", log.Lmicroseconds),
	)

	go influxdb.InfluxDB(metrics.DefaultRegistry,
		10*time.Second,
		"http://127.0.0.1:8086",
		"monitorss",
		"",
		"",
		// "username",
		// "password",
	)

	err := startMonitor(configFile)
	// err := checkGoogle()
	if err != nil {
		log.Fatalln("err", err)
	}
}
