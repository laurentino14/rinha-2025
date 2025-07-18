package handlers

import (
	"context"
	"net/http"

	"github.com/laurentino14/rinha-2025/internal/application/payment"
	"github.com/valyala/fasthttp"
)

func Payment(ctx *fasthttp.RequestCtx, process payment.ProcessUseCase) {
	receivedBody := ctx.Request.Body()
	ctx.SetStatusCode(http.StatusOK)
	bodyCopy := make([]byte, len(receivedBody))
	copy(bodyCopy, receivedBody)
	process(context.Background(), bodyCopy)
}
