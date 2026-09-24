package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"messenger-svyaz/internal/config"
	"messenger-svyaz/internal/db"
	"messenger-svyaz/internal/db/sqlc"
	"messenger-svyaz/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Initialize database
	database, err := db.New(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize sqlc queries
	queries := sqlc.New(database.Pool)

	// Initialize repositories
	keyRepo := repository.NewKeyRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		generateKeys(ctx, keyRepo)
	case "list-keys":
		listKeys(ctx, keyRepo)
	case "ban-user":
		banUser(ctx, userRepo)
	case "stats":
		showStats(ctx, userRepo)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: admin <command> [args]")
	fmt.Println("\nCommands:")
	fmt.Println("  generate -n <count>    Generate registration keys")
	fmt.Println("  list-keys               List all registration keys")
	fmt.Println("  ban-user <username>     Ban/unban a user")
	fmt.Println("  stats                   Show server statistics")
}

func generateKeys(ctx context.Context, keyRepo *repository.KeyRepository) {
	count := flag.Int("n", 1, "Number of keys to generate")
	flag.Parse()

	keys, err := keyRepo.GenerateAndCreate(ctx, *count)
	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	fmt.Printf("Generated %d keys:\n", len(keys))
	for _, key := range keys {
		fmt.Println(key)
	}
}

func listKeys(ctx context.Context, keyRepo *repository.KeyRepository) {
	keys, err := keyRepo.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list keys: %v", err)
	}

	fmt.Printf("Registration keys (%d):\n", len(keys))
	for _, key := range keys {
		status := "active"
		if !key.IsActive {
			status = "inactive"
		}
		usedBy := ""
		if key.UsedBy.Valid {
			usedBy = key.UsedBy.String
		}
		fmt.Printf("  %s - %s - used by: %s\n", key.Key, status, usedBy)
	}
}

func banUser(ctx context.Context, userRepo *repository.UserRepository) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: admin ban-user <username>")
		os.Exit(1)
	}

	username := os.Args[2]

	user, err := userRepo.GetByUsername(ctx, username)
	if err != nil {
		log.Fatalf("User not found: %v", err)
	}

	// Toggle ban status
	newStatus := !user.IsBlocked
	err = userRepo.UpdateBlocked(ctx, username, newStatus)
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}

	status := "unbanned"
	if newStatus {
		status = "banned"
	}
	fmt.Printf("User %s has been %s\n", username, status)
}

func showStats(ctx context.Context, userRepo *repository.UserRepository) {
	users, err := userRepo.ListAll(ctx)
	if err != nil {
		log.Fatalf("Failed to get users: %v", err)
	}

	onlineCount := 0
	adminCount := 0
	blockedCount := 0

	for _, user := range users {
		if user.IsOnline {
			onlineCount++
		}
		if user.IsAdmin {
			adminCount++
		}
		if user.IsBlocked {
			blockedCount++
		}
	}

	fmt.Println("Server Statistics:")
	fmt.Printf("  Total users: %d\n", len(users))
	fmt.Printf("  Online users: %d\n", onlineCount)
	fmt.Printf("  Admin users: %d\n", adminCount)
	fmt.Printf("  Blocked users: %d\n", blockedCount)
}
