package main

import (
	"context"
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
	// #region agent log H6
	f, _ := os.OpenFile("/home/user/git/ai-for-developers-project-386/.cursor/debug-eb49d8.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, `{"sessionId":"eb49d8","runId":"fix-verify","hypothesisId":"H6","location":"main.go:start","message":"Server starting","data":{},"timestamp":%d}`+"\n", time.Now().UnixMilli())
		f.Close()
	}
	// #endregion

	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found or error loading: %v", err)
	}

	// #region agent log H6
	f, _ = os.OpenFile("/home/user/git/ai-for-developers-project-386/.cursor/debug-eb49d8.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		dsn := os.Getenv("DATABASE_URL")
		port := os.Getenv("PORT")
		fmt.Fprintf(f, `{"sessionId":"eb49d8","runId":"fix-verify","hypothesisId":"H6","location":"main.go:after_env_load","message":"Environment loaded","data":{"dsn_empty":"%v","dsn_preview":"%s","port":"%s"},"timestamp":%d}`+"\n", dsn == "", dsn[:min(len(dsn), 30)], port, time.Now().UnixMilli())
		f.Close()
	}
	// #endregion

	ctx := context.Background()

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Run migrations with error handling for idempotent operations
	if err := db.Migrate(ctx, pool, "migrations"); err != nil {
		// Log error but don't fail - tables may already exist
		fmt.Printf("Migration warning: %v\n", err)
		fmt.Println("Continuing startup (tables should already exist)...")
	}

	router := api.NewRouter(pool)

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
