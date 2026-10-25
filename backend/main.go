package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"spaceedu/db"
	"spaceedu/handlers"
	"spaceedu/middleware"
)

func main() {
	godotenv.Load(".env")

	if err := db.Connect(); err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	defer db.Pool.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("../"))
	mux.Handle("/", fs)

	mux.HandleFunc("/api/auth/register", methodOnly("POST", handlers.Register))
	mux.HandleFunc("/api/auth/login", methodOnly("POST", handlers.Login))
	mux.Handle("/api/auth/me", middleware.Auth(http.HandlerFunc(handlers.Me)))

	mux.HandleFunc("/api/courses", handlers.ListCourses)
	mux.HandleFunc("/api/courses/", handlers.GetCourse)

	mux.Handle("/api/me/enrollments", middleware.Auth(http.HandlerFunc(handlers.ListMyEnrollments)))
	mux.Handle("/api/checkout", middleware.Auth(http.HandlerFunc(handlers.Checkout)))
	mux.Handle("/api/checkout/", middleware.Auth(http.HandlerFunc(handlers.GetOrder)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := corsMiddleware(mux)

	fmt.Printf("SpaceEdu server running on http://localhost:%s\n", port)
	fmt.Printf("Frontend: http://localhost:%s/index.html\n", port)
	fmt.Printf("API:      http://localhost:%s/api/\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func methodOnly(m string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != m {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := []string{
			"http://localhost:8080", "http://127.0.0.1:8080",
			"http://localhost:5500", "http://127.0.0.1:5500",
			"http://localhost:3000",
		}
		for _, a := range allowed {
			if strings.EqualFold(origin, a) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
