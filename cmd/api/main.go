package main

import (
	"L3.4/migrations"
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
	"github.com/wb-go/wbf/retry"
	"os"
	"os/signal"
	"syscall"
	"time"

	"L3.4/internal/config"
	"L3.4/internal/handler"
	"L3.4/internal/logger"
	"L3.4/internal/repository"
	"L3.4/internal/service"
	"L3.4/internal/storage"
	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/ginext"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.Init(cfg)
	if err != nil {
		panic(err)
	}

	sqlDB, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Error("failed to open sql db for migrations", "error", err)
		log.Error("Migration DB connection failed")
		os.Exit(1)
	}
	if err := migrations.Apply(sqlDB); err != nil {
		log.Error("failed to apply migrations", "error", err)
		log.Error("Migrations failed")
	}
	sqlDB.Close()

	var db *pgxdriver.Postgres
	strategy := retry.Strategy{
		Attempts: 10,
		Delay:    1 * time.Second,
		Backoff:  2.0,
	}
	err = retry.Do(func() error {
		db, err = pgxdriver.New(cfg.PostgresDSN, log,
			pgxdriver.MaxPoolSize(20),
			pgxdriver.MaxConnAttempts(1),
		)
		return err
	}, strategy)
	if err != nil {
		log.Error("Failed to connect to database after retries", "error", err)
		log.Error("DB connection failed")
	}
	defer db.Close()

	repo := repository.NewPostgresRepository(db, log)

	st, err := storage.NewLocalStorage(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", "error", err)
		os.Exit(1)
	}

	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopic,
		Balancer: &kafka.LeastBytes{},
	})
	defer kafkaWriter.Close()

	svc := service.NewImageService(repo, st, kafkaWriter, log, cfg.StoragePath)

	router := ginext.New("debug")
	router.Use(ginext.Logger(), ginext.Recovery())

	h := handler.NewHandler(svc, log)
	h.RegisterRoutes(router)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := router.Run(":" + cfg.HTTPPort); err != nil {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down gracefully")
}
