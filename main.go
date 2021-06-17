package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/zacscoding/burrow-all/internal/httpserver"
	"github.com/zacscoding/burrow-all/internal/kafka"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var (
		kafkaBrokers = []string{
			"localhost:9092",
		}
		kafkaVersion = "2.2.0"
		serverAddr   = ":8900"
	)

	e := echo.New()
	srv := httpserver.NewServer(e, &kafka.Config{
		Brokers:      kafkaBrokers,
		KafkaVersion: kafkaVersion,
	})
	go func() {
		if err := srv.Start(serverAddr); err != nil && err != http.ErrServerClosed {
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
