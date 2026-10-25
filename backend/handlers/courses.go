package handlers

import (
	"context"
	"net/http"
	"strings"

	"spaceedu/db"
	"spaceedu/middleware"
	"spaceedu/models"
)

func ListCourses(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, slug, title, description, planet, price, duration_hours, image_url, created_at
		 FROM courses ORDER BY planet, created_at`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not fetch courses"})
		return
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var c models.Course
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Description, &c.Planet,
			&c.Price, &c.DurationHours, &c.ImageURL, &c.CreatedAt); err != nil {
			continue
		}
		courses = append(courses, c)
	}

	writeJSON(w, http.StatusOK, courses)
}

func GetCourse(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/courses/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing course id"})
		return
	}

	var c models.Course
	err := db.Pool.QueryRow(context.Background(),
		`FROM courses WHERE id::text=$1 OR slug=$1`, id,
	).Scan(&c.ID, &c.Slug, &c.Title, &c.Description, &c.Planet,
		&c.Price, &c.DurationHours, &c.ImageURL, &c.CreatedAt)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "course not found"})
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func ListMyEnrollments(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	rows, err := db.Pool.Query(context.Background(),
		`SELECT e.id, e.user_id, e.course_id, COALESCE(e.order_id::text,''), e.enrolled_at,
		        c.id, c.slug, c.title, c.description, c.planet, c.price, c.duration_hours, c.image_url, c.created_at
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.user_id=$1
		 ORDER BY e.enrolled_at DESC`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not fetch enrollments"})
		return
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		var c models.Course
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.CourseID, &e.OrderID, &e.EnrolledAt,
			&c.ID, &c.Slug, &c.Title, &c.Description, &c.Planet,
			&c.Price, &c.DurationHours, &c.ImageURL, &c.CreatedAt,
		); err != nil {
			continue
		}
		e.Course = &c
		enrollments = append(enrollments, e)
	}

	if enrollments == nil {
		enrollments = []models.Enrollment{}
	}
	writeJSON(w, http.StatusOK, enrollments)
}
