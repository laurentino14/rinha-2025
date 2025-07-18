package payment

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"

	"github.com/laurentino14/rinha-2025/internal/application/models"

	"github.com/laurentino14/rinha-2025/internal/config"
)

const (
	ResultInvalidInput = iota - 1
	ResultRetry
	ResultSuccess
	ResultProcessorOffline
)
const (
	CommonQueue = "payment"
)

type ProcessWorkerPool struct {
	cache     *cache
	paymentDB Repository
	queue     chan []byte
	client    *fasthttp.Client
	ctx       context.Context
}

func NewProcessWorkerPool(paymentDB Repository, cache *cache) *ProcessWorkerPool {
	return &ProcessWorkerPool{
		paymentDB: paymentDB,
		cache:     cache,
		queue:     make(chan []byte, 10000),
		ctx:       context.Background(),
		client: &fasthttp.Client{
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			MaxConnsPerHost: 1000,
		},
	}
}

func (s *ProcessWorkerPool) Worker() {
	for data := range s.queue {
		var retry int
		var msg models.Payment
		json.Unmarshal(data, &msg)
	MessageProcessingLoop:
		for {
			s.cache.M.RLock()
			active, lastChecked := s.cache.ActiveProcessor, s.cache.LastChecked
			s.cache.M.RUnlock()

			var processorID int
			var processorURL string
			var result int

			switch active {
			case 0:
				secondsToWait := 5*time.Second - time.Since(lastChecked)
				if secondsToWait < 0 {
					secondsToWait = 100 * time.Millisecond
				}
				time.Sleep(secondsToWait)
				continue
			case 1:
				processorID = 1
				processorURL = config.DefaultURL + "/payments"

			case 2:
				processorID = 2
				processorURL = config.FallbackURL + "/payments"
			}

			result = s.SendToProcessor(processorURL, processorID, &msg, 1)

			switch result {
			case ResultSuccess:
				break MessageProcessingLoop
			case ResultInvalidInput:
				slog.Error("[Priority Worker]: Invalid Input")
				break MessageProcessingLoop
			case ResultRetry:
				retry++
				if retry > 3 {
					slog.Error("[Priority Worker]: Retry Excedded")
					break MessageProcessingLoop
				}
				secondsToWait := 5*time.Second - time.Since(lastChecked)
				if secondsToWait < 0 {
					secondsToWait = 100 * time.Millisecond
				}
				time.Sleep(secondsToWait)
				continue
			case ResultProcessorOffline:
				time.Sleep(100 * time.Millisecond)
				continue
			}

		}
	}
}

func (s *ProcessWorkerPool) SendMessage(data []byte) {
	select {
	case s.queue <- data:
		return
	default:
		return
	}
}

func (s *ProcessWorkerPool) SendToProcessor(processorURL string, processorID int, msg *models.Payment, queue int) int {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)
	var duration time.Duration
	if processorID == 1 {
		duration = time.Second * 2
	} else {
		duration = time.Second * 3
	}
	msg.RequestedAt = time.Now().UTC().Format(time.RFC3339Nano)
	body, _ := json.Marshal(msg)
	req.SetBody(body)
	req.SetRequestURI(processorURL)
	req.Header.SetMethod(http.MethodPost)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Rinha-Token", config.Token)

	err := s.client.DoTimeout(req, res, duration)
	status := res.StatusCode()

	if err != nil || status < http.StatusOK || status >= http.StatusMultipleChoices {
		if status == http.StatusUnprocessableEntity {
			return ResultInvalidInput
		}
		return ResultRetry
	}

	s.paymentDB.SavePayment(s.ctx, msg, processorID)

	return ResultSuccess
}

func (s *ProcessWorkerPool) Run(workers int) {
	for range workers {
		go s.Worker()
	}
}
