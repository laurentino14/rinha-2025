package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/laurentino14/rinha-2025/internal/application/payment"
	"github.com/valyala/fasthttp"
)

func GetSummary(ctx *fasthttp.RequestCtx, usecase payment.GetSummaryUseCase) {
	args := ctx.QueryArgs()
	from, to := string(args.Peek("from")), string(args.Peek("to"))
	r := usecase(ctx, from, to)
	b, _ := json.Marshal(r)
	ctx.SetBody(b)
	ctx.SetStatusCode(http.StatusOK)
}
