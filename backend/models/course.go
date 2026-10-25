package models

import "time"

type Course struct {
	ID            string    `json:"id"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Planet        string    `json:"planet"`
	Price         float64   `json:"price"`
	DurationHours int       `json:"duration_hours"`
	ImageURL      string    `json:"image_url"`
	CreatedAt     time.Time `json:"created_at"`
}
