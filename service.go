package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/rcrowley/go-metrics"
)

const (
	// restart how often
	restartTimeout = 10 * time.Second
	// how long to wait for process to startup before starting to check
	startupTime = 10 * time.Second
	// check google how often
	checkTimeout = 10 * time.Second
)

type SSServerConfig struct {
	Server     string
	ServerPort int `json:"server_port"`
	Password   string
	Method     string
	Timeout    int
	FastOpen   bool `json:"fast_open"`
}

type ssService struct {
	id string
	// index     int
	localPort int
	// name      string
	config SSServerConfig

	stat ssServiceStat
}

type ssServiceStat struct {
	Fails            uint64
	OKs              uint64
	LastPingDuration time.Duration
}

func (s *ssService) start(ctx context.Context) {
	for {
		err := s.run(ctx)

		if err != nil {
			log.Println("Service ended. Restart in 3 seconds. Reason:", err)
		}

		time.Sleep(restartTimeout)
	}
}

func (s *ssService) checkLoop(ctx context.Context, cancel context.CancelFunc) {

	port := s.localPort

	googleURL := fmt.Sprintf("https://google.com?port=%d", port)

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(fmt.Sprintf("socks5://localhost:%d", port))
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	for {
		// check if process is already dead...
		select {
		case <-ctx.Done():
			break
		default:
		}

		ping := checkURL(client, googleURL)

		if ping.Err != nil {
			cancel()
			break
		}

		// if ping.Err != nil {
		// 	s.stat.Fails++
		// } else {
		// 	s.stat.OKs++
		// }
		// s.stat.LastPingDuration = ping.Duration

		t := metrics.GetOrRegisterTimer(fmt.Sprintf("%s.ping", s.id), nil)
		t.Update(ping.Duration)

		// log.Printf("[%s] %v\n", s.id, s.stat)

		stagger := rand.Intn(300)
		waitTime := checkTimeout + time.Duration(stagger)*time.Millisecond
		time.Sleep(waitTime)
	}
}

func (s *ssService) run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)

	config := s.config

	cmdArgs := []string{
		"-s", config.Server,
		"-p", strconv.Itoa(config.ServerPort),
		"-l", strconv.Itoa(s.localPort),
		"-k", config.Password,
		"-m", config.Method,
	}

	log.Printf("[%s]: ss-local %v\n", s.id, cmdArgs)

	cmd := exec.CommandContext(ctx, "ss-local", cmdArgs...)

	cmd.Stderr = os.Stderr

	go func() {
		time.Sleep(5 * time.Second)
		// cancel the command if check fails
		s.checkLoop(ctx, cancel)
	}()

	err = cmd.Run()

	// cancel the check loop if process finished
	cancel()

	if err != nil {
		return
	}

	return nil
}
