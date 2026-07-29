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

// Package config loads application configuration from environment variables
// (and an optional .env file) into a single, process-wide Config value.
//
// Call Load once during startup, before any other package reads Cfg.
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds every tunable the worker needs at runtime: RabbitMQ
// connection details, consumer behavior, archive output location, and
// logging options. It is populated exclusively by Load.
type Config struct {
	// RabbitConsumeQueue is the name of the (dead-letter) queue to consume from.
	RabbitConsumeQueue string

	// WorkerCount is the number of concurrent message handlers (RB_WORKER_COUNT).
	WorkerCount int
	// AMQPReconnectDelay is the delay between reconnect attempts (AMQPR_CON_DELAY).
	AMQPReconnectDelay time.Duration

	// RabbitHost is the broker hostname or IP (RB_HOST).
	RabbitHost string
	// RabbitUsername is the broker login username (RB_USERNAME).
	RabbitUsername string
	// RabbitPassword is the broker login password (RB_PASSWD).
	RabbitPassword string
	// RabbitPort is the broker AMQP port (RB_PORT, default 5672).
	RabbitPort int
	// RabbitVHost is the broker virtual host (RB_VHOST, default "/").
	RabbitVHost string
	// RabbitPrefetch is the per-consumer QoS prefetch count (RB_PREFETCH).
	RabbitPrefetch int

	// ConsumerTag identifies this consumer to the broker (RB_CONSUMER_TAG).
	ConsumerTag string

	// ArchiveBasePath is the root directory archived messages are written under (ARCHIVE_BASE_PATH).
	ArchiveBasePath string

	// LogFile is the path of the rotating log file (LOG_FILE).
	LogFile string
	// LogMaxSizeMB is the size in megabytes at which the log file is rotated (LOG_MAX_SIZE_MB).
	LogMaxSizeMB int
	// LogMaxBackups is the number of rotated log files to retain (LOG_MAX_BACKUPS).
	LogMaxBackups int
	// LogMaxAgeDays is the number of days to retain rotated log files (LOG_MAX_AGE_DAYS).
	LogMaxAgeDays int
	// LogCompress controls whether rotated log files are gzip-compressed (LOG_COMPRESS).
	LogCompress bool
	// LogLevel is the minimum zerolog level to emit, e.g. "debug", "info" (LOG_LEVEL).
	LogLevel string
	// LogShowConsole controls whether logs are also written to stderr (LOG_CONSOLE).
	LogShowConsole bool
}

// Cfg is the global, process-wide configuration instance. It is nil until
// Load is called and must not be read before then.
var Cfg *Config

// getenv returns the environment variable named key, or def if it is unset
// or empty.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// getint returns the environment variable named key parsed as an int, or
// def if it is unset, empty, or not a valid integer.
func getint(key string, def int) int {
	v := getenv(key, "")
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

// getbool returns the environment variable named key parsed as a bool, or
// def if it is unset, empty, or not a valid boolean (as accepted by
// strconv.ParseBool).
func getbool(key string, def bool) bool {
	v := getenv(key, "")
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

// getdur returns the environment variable named key parsed as a
// time.Duration (e.g. "5s", "1m"), or def if it is unset, empty, or not a
// valid duration.
func getdur(key string, def time.Duration) time.Duration {
	v := getenv(key, "")
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

// Load reads a .env file if present (missing files are silently ignored),
// then populates the global Cfg from environment variables, applying the
// defaults documented on each Config field. It must be called once before
// any other package reads Cfg, and is not safe to call concurrently with
// reads of Cfg.
func Load() {
	_ = godotenv.Load()
	Cfg = &Config{
		RabbitConsumeQueue: getenv("RB_CONSUME_QUEUE", ""),

		WorkerCount:    getint("RB_WORKER_COUNT", 1),
		RabbitPrefetch: getint("RB_PREFETCH", 1),

		RabbitHost:     getenv("RB_HOST", ""),
		RabbitUsername: getenv("RB_USERNAME", ""),
		RabbitPassword: getenv("RB_PASSWD", ""),
		RabbitPort:     getint("RB_PORT", 5672),
		RabbitVHost:    getenv("RB_VHOST", "/"),

		ConsumerTag:        getenv("RB_CONSUMER_TAG", ""),
		AMQPReconnectDelay: getdur("AMQPR_CON_DELAY", 5*time.Second),

		ArchiveBasePath: getenv("ARCHIVE_BASE_PATH", "./archive"),

		LogFile:        getenv("LOG_FILE", "./logs/app.log"),
		LogMaxSizeMB:   getint("LOG_MAX_SIZE_MB", 50),
		LogMaxBackups:  getint("LOG_MAX_BACKUPS", 5),
		LogMaxAgeDays:  getint("LOG_MAX_AGE_DAYS", 14),
		LogCompress:    getbool("LOG_COMPRESS", true),
		LogLevel:       getenv("LOG_LEVEL", "debug"),
		LogShowConsole: getbool("LOG_CONSOLE", false),
	}
}
