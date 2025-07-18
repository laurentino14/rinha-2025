package payment

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"

	"gopkg.in/zeromq/goczmq.v4"

	"github.com/cloudwego/hertz/pkg/app/client"

	"github.com/laurentino14/rinha-2025/internal/config"
)

type processorStats struct {
	IsFailing       bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

type worker struct {
	cache  *cache
	client *client.Client
	pub    *goczmq.Sock
	sub    *goczmq.Sock
	ctx    context.Context
}

func NewHealthCheckWorker(ctx context.Context, cache *cache) *worker {
	client, _ := client.NewClient()
	var err error
	var pub, sub *goczmq.Sock
	if config.GetDefaultEnv("SERVER", "") == "" {
		sub, err = goczmq.NewSub(config.GetDefaultEnv("SUB_URL", "tcp://api1:5555"), "")
		if err != nil {
			log.Fatal(err)
		}
	}
	if config.GetDefaultEnv("SERVER", "") == "main" {
		pub, err = goczmq.NewPub("tcp://*:5555")
		if err != nil {
			log.Fatal(err)
		}
	}

	return &worker{
		ctx:    ctx,
		cache:  cache,
		client: client,
		pub:    pub,
		sub:    sub,
	}
}

func (s *worker) Run() {

	if config.GetDefaultEnv("SERVER", "") == "" {
		go s.healthListener()
	}
	if config.GetDefaultEnv("SERVER", "") == "main" {
		go s.healthWorker()
	}

}

func (s *worker) healthListener() {
	for {
		msg, _, err := s.sub.RecvFrame()
		if err != nil {
			continue
		}
		code, _ := strconv.Atoi(string(msg))
		if s.cache.ActiveProcessor != code {
			s.cache.M.Lock()
			s.cache.ActiveProcessor = code
			s.cache.LastChecked = time.Now()
			s.cache.M.Unlock()
			switch code {
			case 1:
				slog.Info("[Healthcheck]: changed to DEFAULT")
			case 2:
				slog.Info("[Healthcheck]: changed to FALLBACK")
			}
		} else {
			slog.Info("[Healthcheck]: no changes")
		}

	}
}

func (s *worker) healthWorker() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		dch, fch := make(chan *processorStats), make(chan *processorStats)
		go s.getProcessorStats(config.DefaultURL+"/payments/service-health", dch)
		go s.getProcessorStats(config.FallbackURL+"/payments/service-health", fch)
		code := s.calculate(<-dch, <-fch)
		close(dch)
		close(fch)
		if code != s.cache.ActiveProcessor {
			s.cache.M.Lock()
			s.cache.ActiveProcessor = code
			s.cache.LastChecked = time.Now()
			s.cache.M.Unlock()
			switch code {
			case 1:
				slog.Info("[Healthcheck]: changed to DEFAULT")
			case 2:
				slog.Info("[Healthcheck]: changed to FALLBACK")
			}
		} else {
			slog.Info("[Healthcheck]: no changes")
		}
		s.pub.SendFrame([]byte(strconv.Itoa(code)), goczmq.FlagNone)
	}
}

func (s *worker) calculate(d *processorStats, f *processorStats) int {
	if d.IsFailing && f.IsFailing {
		return 0
	}
	if d.IsFailing && !f.IsFailing {
		return 2
	}
	if !d.IsFailing && f.IsFailing {
		return 1
	}

	// latencyWeigth := 0.3
	preferredWeigth := 0.1

	dScore := float64(d.MinResponseTime) - preferredWeigth
	fScore := float64(f.MinResponseTime)

	if dScore > fScore {
		return 2
	}

	return 1
}

func (s *worker) getProcessorStats(url string, ch chan<- *processorStats) {
	var data processorStats
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURI(url)
	req.Header.SetMethod(http.MethodGet)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Rinha-Token", config.Token)
	err := fasthttp.Do(req, res)
	status := res.StatusCode()
	if err != nil || status < http.StatusOK || status >= http.StatusMultipleChoices {
		data = processorStats{
			IsFailing: true,
		}
		ch <- &data
		return
	}

	if err := json.Unmarshal(res.Body(), &data); err != nil {
		data = processorStats{
			IsFailing: true,
		}
	}
	ch <- &data
}
