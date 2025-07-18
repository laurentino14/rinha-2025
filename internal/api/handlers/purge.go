package handlers

import (
	"net/http"

	"github.com/laurentino14/rinha-2025/internal/application/payment"
	"github.com/valyala/fasthttp"
)

func Purge(ctx *fasthttp.RequestCtx, usecase payment.PurgePaymentsUseCase) {
	usecase(ctx)
	ctx.SetStatusCode(http.StatusOK)
}
