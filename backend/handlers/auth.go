package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"spaceedu/db"
	"spaceedu/middleware"
	"spaceedu/models"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	body.Name = strings.TrimSpace(body.Name)

	if body.Name == "" || body.Email == "" || len(body.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, email and password (min 6 chars) are required"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	var user models.User
	err = db.Pool.QueryRow(context.Background(),
		`INSERT INTO users (name, email, password_hash) VALUES ($1,$2,$3) RETURNING id, name, email, created_at`,
		body.Name, body.Email, string(hash),
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create user"})
		return
	}

	token, err := signToken(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not sign token"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "user": user})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	body.Email = strings.ToLower(strings.TrimSpace(body.Email))

	var user models.User
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, name, email, password_hash, created_at FROM users WHERE email=$1`, body.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	token, err := signToken(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not sign token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

func Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var user models.User
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, name, email, created_at FROM users WHERE id=$1`, userID,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func signToken(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "spaceedu-super-secret-jwt-key-change-in-production"
	}

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
