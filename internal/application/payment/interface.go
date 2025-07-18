package payment

import (
	"context"

	"github.com/laurentino14/rinha-2025/internal/application/models"
)

type writer interface {
	SavePayment(ctx context.Context, payload *models.Payment, processor int)
	Purge(ctx context.Context)
}

type reader interface {
	GetSummary(ctx context.Context, from, to string) models.Summary
}

type Repository interface {
	writer
	reader
}

type ProcessUseCase func(ctx context.Context, data []byte)
type GetSummaryUseCase func(ctx context.Context, from string, to string) models.Summary
type PurgePaymentsUseCase func(ctx context.Context)
