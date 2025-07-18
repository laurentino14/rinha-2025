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

// func (s *ProcessWorkerPool) Worker() {
// 	workerCtx, stop := signal.NotifyContext(s.ctx, os.Interrupt, syscall.SIGTERM)
// 	defer stop()
// 	s.nc.QueueSubscribe("payment", "worker", func(msg *nats.Msg) {
// 		s.cache.M.RLock()
// 		active := s.cache.ActiveProcessor
// 		s.cache.M.RUnlock()
// 		var processorURL string
// 		var processorID int
// 		switch active {
// 		case 1:
// 			processorID = 1
// 			processorURL = config.DefaultURL + "/payments"
// 		case 2:
// 			processorID = 2
// 			processorURL = config.FallbackURL + "/payments"
// 		default:
// 			s.SendToPriorityQueue(msg)
// 			return
// 		}

// 		result := s.SendToProcessor(processorURL, processorID, msg, 0)

// 		if result == ResultRetry {
// 			s.SendToPriorityQueue(msg)
// 		}
// 	})
// 	<-workerCtx.Done()
// }

// func (s *ProcessWorkerPool) PriorityWorker() {
// 	workerCtx, stop := signal.NotifyContext(s.ctx, os.Interrupt, syscall.SIGTERM)
// 	defer stop()
// 	s.nc.QueueSubscribe("payment.priority", "worker", func(msg *nats.Msg) {
// 		var retry int

// 		for {
// 			s.cache.M.RLock()
// 			active, lastChecked := s.cache.ActiveProcessor, s.cache.LastChecked
// 			s.cache.M.RUnlock()

// 			var processorID int
// 			var processorURL string
// 			var result int

// 			switch active {
// 			case 0:
// 				secondsToWait := 5*time.Second - time.Since(lastChecked)
// 				if secondsToWait < 0 {
// 					secondsToWait = 100 * time.Millisecond
// 				}
// 				time.Sleep(secondsToWait)
// 				continue
// 			case 1:
// 				processorID = 1
// 				processorURL = config.DefaultURL + "/payments"

// 			case 2:
// 				processorID = 2
// 				processorURL = config.FallbackURL + "/payments"
// 			}

// 			result = s.SendToProcessor(processorURL, processorID, msg, 1)

// 			switch result {
// 			case ResultSuccess:
// 				return
// 			case ResultInvalidInput:
// 				slog.Error("[Priority Worker]: Invalid Input", slog.String("data", string(msg.Data)))
// 				return
// 			case ResultRetry:
// 				retry++
// 				if retry > 3 {
// 					slog.Error("[Priority Worker]: Retry Excedded", slog.String("data", string(msg.Data)))
// 					return
// 				}
// 				secondsToWait := 5*time.Second - time.Since(lastChecked)
// 				if secondsToWait < 0 {
// 					secondsToWait = 100 * time.Millisecond
// 				}
// 				time.Sleep(secondsToWait)
// 				continue
// 			case ResultProcessorOffline:
// 				time.Sleep(100 * time.Millisecond)
// 				continue
// 			}

// 		}
// 	})
// 	<-workerCtx.Done()
// }

// func (s *ProcessWorkerPool) Subscriber(){
// 	subscriberCtx, stop := signal.NotifyContext(s.ctx, os.Interrupt, syscall.SIGTERM)
// 	defer stop()
// 		// s.nc.QueueSubscribe("payment", "worker", func(msg *nats.Msg) {
// 		// 	var payment models.Payment
// 		// 	if err := json.Unmarshal(msg.Data, &payment); err != nil {
// 		// 		return
// 		// 	}
// 		// 	s.queue <- &payment
// 		// })
// 	<- subscriberCtx.Done()
// }

func (s *ProcessWorkerPool) Worker() {
	// workerCtx, stop := signal.NotifyContext(s.ctx, os.Interrupt, syscall.SIGTERM)
	// defer stop()

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

	// <-workerCtx.Done()
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

// func (s *ProcessWorkerPool) SendToProcessor(processorURL string, processorID int, msg *nats.Msg, queue int) int {
// 	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
// 	defer fasthttp.ReleaseRequest(req)
// 	defer fasthttp.ReleaseResponse(res)
// 	var payment *models.Payment
// 	var duration time.Duration
// 	if err := json.Unmarshal(msg.Data, &payment); err != nil {
// 		return ResultInvalidInput
// 	}
// 	if processorID == 1 {
// 		duration = time.Second * 2
// 	}else{
// 		duration = time.Second * 5
// 	}
// 	payment.RequestedAt = time.Now().UTC().Format(time.RFC3339Nano)
// 	body, _ := json.Marshal(payment)
// 	req.SetBody(body)
// 	req.SetRequestURI(processorURL)
// 	req.Header.SetMethod(http.MethodPost)
// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("X-Rinha-Token", config.Token)

// 	err := s.client.DoTimeout(req, res, duration)
// 	status := res.StatusCode()

// 	if err != nil || status < http.StatusOK || status >= http.StatusMultipleChoices {
// 		if status == http.StatusUnprocessableEntity {
// 			return ResultInvalidInput
// 		}
// 		return ResultRetry
// 	}

// 	s.paymentDB.SavePayment(s.ctx,payment, processorID)

// 	return ResultSuccess
// }

// func (s *ProcessWorkerPool) SendToPriorityQueue(msg *nats.Msg) {
// 	s.queue.PublishMsg(&nats.Msg{
// 		Subject: PriorityQueue,
// 		Data:    msg.Data,
// 	})
// }

func (s *ProcessWorkerPool) Run(workers int) {
	for range workers {
		go s.Worker()
	}
}
