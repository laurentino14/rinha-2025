package models

type Payment struct {
	CorrelationID string  `json:"correlationId" redis:"correlation_id"`
	Amount        float64 `json:"amount" redis:"amount"`
	RequestedAt   string  `json:"requestedAt" redis:"requested_at"`
}
