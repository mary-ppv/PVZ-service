package models

import "time"

type City string

const (
	CityMoscow City = "Москва"
	CitySpb    City = "Санкт-Петербург"
	CityKazan  City = "Казань"
)

type PVZ struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	City      City      `db:"city" json:"city"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
