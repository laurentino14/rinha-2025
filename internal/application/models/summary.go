package models

type SummaryProcessor struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type Summary struct {
	Default  SummaryProcessor `json:"default"`
	Fallback SummaryProcessor `json:"fallback"`
}
