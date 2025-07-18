package pg

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laurentino14/rinha-2025/internal/application/models"
	"github.com/laurentino14/rinha-2025/internal/application/payment"
)

type PaymentRepository struct {
	payment.Repository
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) payment.Repository {
	return &PaymentRepository{
		db: db,
	}
}

func (s *PaymentRepository) SavePayment(ctx context.Context, payload *models.Payment, processor int) {
	_, err := s.db.Exec(ctx, "INSERT INTO payments(correlation_id, amount, processor, requested_at) VALUES ($1, $2, $3, $4) ON CONFLICT (correlation_id) DO NOTHING;", payload.CorrelationID, payload.Amount, processor, payload.RequestedAt)
	if err != nil {
		log.Printf("Erro na entrega do postgres: %v", err)
	}
}

func (s *PaymentRepository) SavePaymentBatch(batch *pgx.Batch) {
	s.db.SendBatch(context.Background(), batch)
}

func (s *PaymentRepository) GetSummary(ctx context.Context, from, to string) models.Summary {

	response := models.Summary{
		Default: models.SummaryProcessor{
			TotalRequests: 0,
			TotalAmount:   0,
		},
		Fallback: models.SummaryProcessor{
			TotalRequests: 0,
			TotalAmount:   0,
		},
	}
	f, t := time.Unix(0, 0).UTC().Format(time.RFC3339Nano), time.Now().UTC().AddDate(0, 0, 1).Format(time.RFC3339Nano)
	if from != "" {
		f = from
	}

	if to != "" {
		t = to
	}

	rows, err := s.db.Query(ctx, "SELECT processor, COUNT(*) AS total, SUM(amount) FROM payments WHERE requested_at BETWEEN $1 AND $2 GROUP BY processor;", f, t)
	if err != nil {
		return response
	}
	defer rows.Close()
	for rows.Next() {
		var processor int
		var total int
		var sum float64
		err := rows.Scan(&processor, &total, &sum)
		if err == nil {
			if processor == 1 {
				response.Default.TotalRequests = total
				// ratio := math.Pow(10, float64(2))
				// response.Default.TotalAmount = math.Round(sum*ratio) / ratio
				response.Default.TotalAmount = sum
			} else {
				response.Fallback.TotalRequests = total
				// ratio := math.Pow(10, float64(2))
				// response.Fallback.TotalAmount = math.Round(sum*ratio) / ratio
				response.Fallback.TotalAmount = sum
			}
		}
	}
	return response
}

func (s *PaymentRepository) Purge(ctx context.Context) {
	s.db.Exec(ctx, "TRUNCATE TABLE payments RESTART IDENTITY;")
}
