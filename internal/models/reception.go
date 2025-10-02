package models

import "time"

type ReceptionStatus string

const (
	ReceptionInProgress ReceptionStatus = "in_progress"
	ReceptionClosed     ReceptionStatus = "closed"
)

type Reception struct {
	ID         string          `json:"id"`
	PvzID      string          `json:"pvzId"`
	DateTime   time.Time       `json:"dateTime"`
	ProductIDs []string        `json:"productIds"` // UUID товаров
	Status     ReceptionStatus `json:"status"`
}
