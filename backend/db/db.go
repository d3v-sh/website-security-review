package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://spaceedu:spaceedu_secret@localhost:5433/spaceedu?sslmode=disable"
	}

	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	if err = Pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("Database connected")
	return nil
}

func Migrate() error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,

		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS courses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			planet TEXT NOT NULL,
			price NUMERIC(10,2) NOT NULL DEFAULT 0,
			duration_hours INT NOT NULL DEFAULT 0,
			image_url TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
			amount NUMERIC(10,2) NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS enrollments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
			order_id UUID REFERENCES orders(id),
			enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(user_id, course_id)
		)`,
	}

	for _, q := range queries {
		_, err := Pool.Exec(context.Background(), q)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	var count int
	Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM courses").Scan(&count)
	if count == 0 {
		if err := seedCourses(); err != nil {
			return fmt.Errorf("seed failed: %w", err)
		}
		log.Println("Courses seeded")
	}

	log.Println("Migrations complete")
	return nil
}

func seedCourses() error {
	courses := []struct {
		slug, title, description, planet, imageURL string
		price                                       float64
		duration                                    int
	}{
		{"earth-101", "Earth 101: Our Living Planet", "Explore Earth's ecosystems, geology, atmosphere, and everything that makes our planet uniquely habitable in the cosmos.", "earth", "https://d8j0ntlcm91z4.cloudfront.net/user_38xzZboKViGWJOttwIXH07lWA1P/hf_20260827_202133_508c64b8-a31e-4290-bdfc-1187df70e0a6.png", 49.99, 12},
		{"earth-climate", "Earth's Climate Systems", "Dive deep into atmospheric science, ocean currents, and climate change. Understand the forces shaping our world's future.", "earth", "https://d8j0ntlcm91z4.cloudfront.net/user_38xzZboKViGWJOttwIXH07lWA1P/hf_20260827_202133_508c64b8-a31e-4290-bdfc-1187df70e0a6.png", 79.99, 18},
		{"venus-atmosphere", "Venus: The Hellish Twin", "Journey to Venus and understand why Earth's nearest neighbor became a runaway greenhouse world, and what it means for our future.", "venus", "https://d8j0ntlcm91z4.cloudfront.net/user_38xzZboKViGWJOttwIXH07lWA1P/hf_20260827_202133_cf55d1d8-7b59-4a64-80da-d72052ae974e.png", 59.99, 10},
		{"venus-exploration", "Exploring Venus: Past & Future Missions", "From Mariner 2 to future atmospheric probes, trace humanity's exploration of our cloud-shrouded sister planet.", "venus", "https://d8j0ntlcm91z4.cloudfront.net/user_38xzZboKViGWJOttwIXH07lWA1P/hf_20260827_202133_cf55d1d8-7b59-4a64-80da-d72052ae974e.png", 69.99, 14},
		{"mars-geology", "Mars: The Red Desert World", "Study Mars's geology, Olympus Mons, Valles Marineris, and the evidence for ancient liquid water on the rust-red desert planet.", "mars", "https://d8j0ntlcm91z4.cloudfront.net/user_38xzZboKViGWJOttwIXH07lWA1P/hf_20260827_202133_0ba6de7c-285d-43dc-b7ab-8c54c73707cb.png", 54.99, 11},
		{"mars-colonization", "Mars Colonization: Science & Strategy", "Explore the engineering, biology, and ethical dimensions of establishing a permanent human presence on Mars.", "mars", "https://d8j0ntlcm91z4.cloudfront.net/user_38xzZboKViGWJOttwIXH07lWA1P/hf_20260827_202133_0ba6de7c-285d-43dc-b7ab-8c54c73707cb.png", 89.99, 20},
	}

	for _, c := range courses {
		_, err := Pool.Exec(context.Background(),
			`INSERT INTO courses (slug, title, description, planet, price, duration_hours, image_url) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			c.slug, c.title, c.description, c.planet, c.price, c.duration, c.imageURL,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
