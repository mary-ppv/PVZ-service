package models

import "time"

type ProductType string

const (
	ProductElectronics ProductType = "electronics"
	ProductClothes     ProductType = "clothes"
	ProductShoes       ProductType = "shoes"
)

type Product struct {
	ID          string      `db:"id" json:"id"`
	ReceptionID string      `db:"reception_id" json:"reception_id"`
	Type        ProductType `db:"type" json:"type"`
	AddedAt     time.Time   `db:"added_at" json:"added_at"`
}
