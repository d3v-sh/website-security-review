package models

import "time"

type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CourseID  string    `json:"course_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Enrollment struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	CourseID   string    `json:"course_id"`
	OrderID    string    `json:"order_id"`
	EnrolledAt time.Time `json:"enrolled_at"`
	Course     *Course   `json:"course,omitempty"`
}
