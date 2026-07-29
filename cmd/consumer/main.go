// Copyright (C) 2026 Yukthi Systems Private Limited
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License version 3
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// version 3 along with this program. If not, see
// <https://www.gnu.org/licenses/>.

// Command consumer runs the long-lived worker that consumes messages from a
// RabbitMQ dead-letter queue and archives them to disk. Configuration is
// read from the environment (and an optional .env file); see
// internal/config for the full list of supported variables.
//
// The worker shuts down gracefully on SIGINT or SIGTERM, letting in-flight
// deliveries finish before exiting.
package main

import (
	"context"
	"fmt"

	"os"
	"os/signal"

	"syscall"

	"github.com/Yukthi-Systems/Archive-Dead-Queue-Worker-RMQ/internal/config"
	"github.com/Yukthi-Systems/Archive-Dead-Queue-Worker-RMQ/internal/logger"
	"github.com/Yukthi-Systems/Archive-Dead-Queue-Worker-RMQ/internal/rmq"
)

// main loads configuration, initializes logging, connects to RabbitMQ, and
// runs the consumer until it is cancelled by a SIGINT/SIGTERM signal.
func main() {
	config.Load()
	rmq_url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		config.Cfg.RabbitUsername,
		config.Cfg.RabbitPassword,
		config.Cfg.RabbitHost,
		config.Cfg.RabbitPort,
		config.Cfg.RabbitVHost,
	)
	logger.Init(logger.Options{
		File:       config.Cfg.LogFile,
		MaxSizeMB:  config.Cfg.LogMaxSizeMB,
		MaxAgeDays: config.Cfg.LogMaxAgeDays,
		MaxBackups: config.Cfg.LogMaxBackups,
		Compress:   config.Cfg.LogCompress,
		Level:      config.Cfg.LogLevel,
		Console:    config.Cfg.LogShowConsole,
	})
	cfg := rmq.Config{
		URL:          rmq_url,
		ConsumeQueue: config.Cfg.RabbitConsumeQueue,
		Prefetch:     config.Cfg.RabbitPrefetch,
		Concurrency:  config.Cfg.WorkerCount,
	}

	conn, err := rmq.NewConnection(cfg)
	if err != nil {
		// log.Fatal(err)
		logger.Fatal().Err(err).Msg("RabbitMQ connection failure.")
	}
	defer conn.Close()

	handler := rmq.NewDefaultHandler(config.Cfg.ArchiveBasePath)

	consumer, err := rmq.NewConsumer(conn, handler, cfg)
	if err != nil {
		// log.Fatal(err)
		logger.Fatal().Err(err).Msg("starting consumer queue failure.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		cancel()
		consumer.Close()
	}()

	// log.Println("Worker started...")
	fmt.Println("worker started...")
	logger.Info().Msg("worker started...")
	if err := consumer.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("starting consumer queue failure.")
	}

}
