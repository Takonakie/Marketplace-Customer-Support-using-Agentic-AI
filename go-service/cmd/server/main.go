package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agenticsdk/go-service/internal/api"
	"github.com/agenticsdk/go-service/internal/bot"
	"github.com/agenticsdk/go-service/internal/config"
	"github.com/agenticsdk/go-service/internal/queue"
	"github.com/agenticsdk/go-service/internal/repository"
)

func main() {
	cfg := config.Load()
	log.Println("Starting Go service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var dbPool = func() *repository.OrderRepo { return nil }()
	_ = dbPool // handle optional db connection logging

	pool, err := config.NewDBPool(cfg.DatabaseURL)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
	} else {
		log.Println("Database connection pool established")
		defer pool.Close()
	}

	rdb, err := queue.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		log.Println("Redis client established")
	}

	var orderRepo *repository.OrderRepo
	var caseRepo *repository.CaseRepo
	var userRepo *repository.UserRepo
	var docRepo *repository.DocRepo

	if pool != nil {
		orderRepo = repository.NewOrderRepo(pool)
		caseRepo = repository.NewCaseRepo(pool)
		userRepo = repository.NewUserRepo(pool)
		docRepo = repository.NewDocRepo(pool)
	}

	router := api.NewRouter(orderRepo, caseRepo, userRepo, docRepo)

	if rdb != nil {
		botService, err := bot.NewBotService(cfg.TelegramBotToken, rdb)
		if err != nil {
			log.Printf("Failed to initialize bot service: %v", err)
		} else {
			go botService.Start(ctx)
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("Go HTTP server listening on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down Go service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
