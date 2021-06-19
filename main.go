package main

import (
	"context"
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/labstack/echo/v4"
	"github.com/zacscoding/burrow-all/internal/httpserver"
	"github.com/zacscoding/burrow-all/internal/kafka"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var cfg config

type config struct {
	KafkaBrokers []string `env:"BURROW_ALL_KAFKA_BROKERS" envSeparator:","`
	KafkaVersion string   `env:"BURROW_ALL_KAFKA_VERSION"`
	ServerAddr   string   `env:"BURROW_ALL_SERVER_ADDR"`
}

func init() {
	brokers := flag.String("kafka-brokers", "localhost:9092,localhost:9093,localhost:9094", "Kafka brokers which comma separated")
	version := flag.String("kafka-version", "2.2.0", "Kafka version")
	addr := flag.String("addr", ":8900", "server address")
	flag.Parse()

	cfg.KafkaBrokers = strings.Split(*brokers, ",")
	cfg.KafkaVersion = *version
	cfg.ServerAddr = *addr

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Printf("Start burrow-all server")
	log.Printf("Kafka brokers: %s", cfg.KafkaBrokers)
	log.Printf("Kafka version: %s", cfg.KafkaVersion)
	log.Printf("Server address: %s", cfg.ServerAddr)

	e := echo.New()
	srv := httpserver.NewServer(e, &kafka.Config{
		Brokers:      cfg.KafkaBrokers,
		KafkaVersion: cfg.KafkaVersion,
	})
	go func() {
		if err := srv.Start(cfg.ServerAddr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	exitChannel := make(chan os.Signal, 1)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-exitChannel

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
