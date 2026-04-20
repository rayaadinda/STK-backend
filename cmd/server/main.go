package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"
	"stk-backend/internal/config"
	"stk-backend/internal/database"
	httpserver "stk-backend/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}

	router := httpserver.NewRouter(cfg, db)

	httpServer := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      router,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
	}

	go func() {
		log.Printf("%s running on port %s (%s)", cfg.App.Name, cfg.App.Port, cfg.App.Env)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	shutdown(httpServer, db, cfg.App.ShutdownTimeout)
}

func shutdown(server *http.Server, db *gorm.DB, timeout time.Duration) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	<-signalChannel
	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("database close skipped: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("database close error: %v", err)
	}

	log.Println("server shutdown complete")
}
