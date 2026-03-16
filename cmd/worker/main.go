package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/wb-go/wbf/retry"
	"os"
	"os/signal"
	"syscall"
	"time"

	"L3.4/internal/config"
	"L3.4/internal/logger"
	"L3.4/internal/repository"
	"L3.4/internal/service"
	"L3.4/internal/worker"
	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/dbpg/pgx-driver"
)

func checkKafkaReady(brokers []string, topic string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		return err
	}
	if len(partitions) == 0 {
		return fmt.Errorf("no partitions for topic %s", topic)
	}
	return nil
}

func connectKafkaWithRetry(brokers []string, topic, groupID string) (*kafka.Reader, error) {
	strategy := retry.Strategy{
		Attempts: 10,
		Delay:    2 * time.Second,
		Backoff:  1.5,
	}

	var reader *kafka.Reader
	err := retry.Do(func() error {
		if err := checkKafkaReady(brokers, topic); err != nil {
			log.Warn().Msgf("kafka not ready: %v", err)
			return err
		}

		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		})

		reader = r
		return nil
	}, strategy)

	if err != nil {
		return nil, err
	}
	log.Info().Msg("successfully connected to kafka")
	return reader, nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.Init(cfg)
	if err != nil {
		panic(err)
	}

	pg, err := pgxdriver.New(cfg.PostgresDSN, log, pgxdriver.MaxPoolSize(20))
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pg.Close()

	repo := repository.NewPostgresRepository(pg, log)

	kafkaReader, err := connectKafkaWithRetry(cfg.KafkaBrokers, cfg.KafkaTopic, "image-workers")
	if err != nil {
		log.Error("failed to connect to kafka after retries", "error", err)
		os.Exit(1)
	}
	defer kafkaReader.Close()

	imageService := service.NewImageService(repo, nil, nil, log, cfg.StoragePath)
	consumer := worker.NewKafkaConsumer(kafkaReader, imageService, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("worker started")
	if err := consumer.Run(ctx); err != nil {
		log.Error("worker stopped with error", "error", err)
		os.Exit(1)
	}
}
