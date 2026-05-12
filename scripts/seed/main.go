package main

import (
	"context"
	"fmt"
	"log"

	"github.com/texef-tech/winzzon-blog/internal/config"
	"github.com/texef-tech/winzzon-blog/internal/db"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	exists, err := queries.AdminExists(ctx, cfg.AdminUsername)
	if err != nil {
		log.Fatalf("failed to check admin: %v", err)
	}

	if exists {
		fmt.Println("admin account already exists, skipping seed")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), 12)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	_, err = queries.CreateAdmin(ctx, sqlc.CreateAdminParams{
		Username:     cfg.AdminUsername,
		PasswordHash: string(hash),
	})
	if err != nil {
		log.Fatalf("failed to create admin: %v", err)
	}

	fmt.Printf("admin account '%s' created successfully\n", cfg.AdminUsername)
}
