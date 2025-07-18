package api

import (
	"net/http"
	"time"

	"github.com/laurentino14/rinha-2025/internal/api/handlers"
	"github.com/laurentino14/rinha-2025/internal/application/payment"

	"github.com/valyala/fasthttp"
)

func Setup(paymentDB payment.Repository, p *payment.ProcessWorkerPool) *fasthttp.Server {

	s := &fasthttp.Server{
		Handler: requestHandler(paymentDB, p),

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,

		Concurrency: 256 * 2,

		ReadBufferSize: 1 * 1024,

		WriteBufferSize: 1 * 1024,
	}

	return s
}

func requestHandler(repo payment.Repository, p *payment.ProcessWorkerPool) fasthttp.RequestHandler {

	return func(ctx *fasthttp.RequestCtx) {
		path, method := string(ctx.Path()), string(ctx.Method())
		switch {
		case path == "/payments" && method == http.MethodPost:
			handlers.Payment(ctx, payment.NewProcessUseCase(p))
		case path == "/payments-summary" && method == http.MethodGet:
			handlers.GetSummary(ctx, payment.NewGetSummaryUseCase(repo))
		case path == "/purge-payments" && method == http.MethodPost:
			handlers.Purge(ctx, payment.NewPurgePaymentUseCase(repo))
		default:
			if path == "/payments" || path == "/payments-summary" || path == "/purge-payments" {
				ctx.SetStatusCode(http.StatusMethodNotAllowed)
				return
			}
			ctx.SetStatusCode(http.StatusNotFound)
		}
	}
}
