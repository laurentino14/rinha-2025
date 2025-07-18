package payment

import (
	"context"

	"github.com/laurentino14/rinha-2025/internal/application/models"
)

func NewProcessUseCase(p *ProcessWorkerPool) ProcessUseCase {
	return func(ctx context.Context, data []byte) {
		p.SendMessage(data)
	}
}

func NewGetSummaryUseCase(repo Repository) GetSummaryUseCase {
	return func(ctx context.Context, from, to string) models.Summary {
		return repo.GetSummary(ctx, from, to)
	}
}

func NewPurgePaymentUseCase(repo Repository) PurgePaymentsUseCase {
	return func(ctx context.Context) {
		repo.Purge(ctx)
	}
}
