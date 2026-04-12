package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"call-booking/internal/api"
	"call-booking/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	ctx := context.Background()

	sqldb, err := db.Open(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func(sqldb *sql.DB) { _ = sqldb.Close() }(sqldb)

	if err := db.Migrate(ctx, sqldb, "migrations"); err != nil {
		log.Fatalf("Migrations failed (fix DATABASE_URL / permissions and restart): %v", err)
	}

	router := api.NewRouter(sqldb)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		fmt.Printf("Server starting on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	fmt.Println("Server stopped")
}
