package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"spaceedu/db"
	"spaceedu/middleware"
	"spaceedu/models"
)

func Checkout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body struct {
		CourseID   string `json:"course_id"`
		CardNumber string `json:"card_number"`
		Expiry     string `json:"expiry"`
		CVV        string `json:"cvv"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if body.CourseID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "course_id required"})
		return
	}

	var course models.Course
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, slug, title, price FROM courses WHERE id::text=$1 OR slug=$1`, body.CourseID,
	).Scan(&course.ID, &course.Slug, &course.Title, &course.Price)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "course not found"})
		return
	}

	var existingID string
	db.Pool.QueryRow(context.Background(),
		`SELECT id FROM enrollments WHERE user_id=$1 AND course_id=$2`, userID, course.ID,
	).Scan(&existingID)
	if existingID != "" {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already enrolled in this course"})
		return
	}

	var order models.Order
	err = db.Pool.QueryRow(context.Background(),
		`INSERT INTO orders (user_id, course_id, amount, status) VALUES ($1, $2, $3, 'completed')
		 RETURNING id, user_id, course_id, amount, status, created_at`,
		userID, course.ID, course.Price,
	).Scan(&order.ID, &order.UserID, &order.CourseID, &order.Amount, &order.Status, &order.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create order"})
		return
	}

	var enrollment models.Enrollment
	err = db.Pool.QueryRow(context.Background(),
		`INSERT INTO enrollments (user_id, course_id, order_id) VALUES ($1, $2, $3)
		 RETURNING id, user_id, course_id, COALESCE(order_id::text,''), enrolled_at`,
		userID, course.ID, order.ID,
	).Scan(&enrollment.ID, &enrollment.UserID, &enrollment.CourseID, &enrollment.OrderID, &enrollment.EnrolledAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create enrollment"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"order": order, "enrollment": enrollment, "course": course})
}

func GetOrder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := strings.TrimPrefix(r.URL.Path, "/api/checkout/")

	var order models.Order
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, user_id, course_id, amount, status, created_at
		 FROM orders WHERE id=$1 AND user_id=$2`, id, userID,
	).Scan(&order.ID, &order.UserID, &order.CourseID, &order.Amount, &order.Status, &order.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
		return
	}

	writeJSON(w, http.StatusOK, order)
}
