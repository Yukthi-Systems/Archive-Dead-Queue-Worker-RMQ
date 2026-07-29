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

// Package logger provides a process-wide, JSON-structured zerolog logger
// backed by a size- and age-based rotating file (via lumberjack).
//
// Call Init once during startup to configure the logger, then use the
// package-level Trace/Debug/Info/Warn/Error/Fatal/Panic functions anywhere
// in the program to emit log events.
package logger

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Options configures the global logger created by Init.
type Options struct {
	// File is the path of the log file to write to and rotate.
	File string
	// MaxSizeMB is the size in megabytes at which the log file is rotated.
	MaxSizeMB int
	// MaxBackups is the number of rotated log files to retain.
	MaxBackups int
	// MaxAgeDays is the number of days to retain rotated log files.
	MaxAgeDays int
	// Compress controls whether rotated log files are gzip-compressed.
	Compress bool
	// Level is the minimum level to log, e.g. "debug", "info", "warn".
	// An invalid or empty value falls back to zerolog.InfoLevel.
	Level string
	// Console, if true, also writes log output to stderr in addition to File.
	Console bool
}

var (
	log *zerolog.Logger
	mu  sync.RWMutex
)

// Init initializes (or reloads) the global logger from opts. Log records
// are written as JSON to a rotating file, and additionally to stderr when
// opts.Console is true. It is safe to call again later to reconfigure the
// logger, but concurrent calls to Init and the shorthand log functions are
// synchronized only against torn reads/writes of the underlying pointer,
// not against each other's ordering.
func Init(opts Options) {
	lj := &lumberjack.Logger{
		Filename:   opts.File,
		MaxSize:    opts.MaxSizeMB,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAgeDays,
		Compress:   opts.Compress,
	}
	var out zerolog.LevelWriter
	if opts.Console {
		// JSON to both file and stderr
		out = zerolog.MultiLevelWriter(os.Stderr, lj)
	} else {
		// JSON only to file
		out = zerolog.MultiLevelWriter(lj)
	}

	level := zerolog.InfoLevel
	if l, err := zerolog.ParseLevel(strings.ToLower(opts.Level)); err == nil {
		level = l
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	// zerolog.TimeFieldFormat = "Mon Jan  2 03:04:05 PM MST 2006"
	zerolog.TimestampFieldName = "time"
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().Local()
	}

	newLogger := zerolog.New(out).Level(level).With().Timestamp().Logger()

	mu.Lock()
	log = &newLogger
	mu.Unlock()

	log.Info().Msg("Logger initialized")
}

// get returns the global logger, lazily falling back to an unconfigured
// stdout logger if Init has not been called yet.
func get() *zerolog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if log == nil {
		tmp := zerolog.New(os.Stdout).With().Timestamp().Logger()
		log = &tmp
	}
	return log
}

// Trace starts a new event with the trace level, using the global logger.
func Trace() *zerolog.Event { return get().Trace() }

// Debug starts a new event with the debug level, using the global logger.
func Debug() *zerolog.Event { return get().Debug() }

// Info starts a new event with the info level, using the global logger.
func Info() *zerolog.Event { return get().Info() }

// Warn starts a new event with the warn level, using the global logger.
func Warn() *zerolog.Event { return get().Warn() }

// Error starts a new event with the error level, using the global logger.
func Error() *zerolog.Event { return get().Error() }

// Fatal starts a new event with the fatal level, using the global logger.
// The os.Exit(1) call happens when the returned event is finalized with
// Msg/Msgf/Send, terminating the process.
func Fatal() *zerolog.Event { return get().Fatal() }

// Panic starts a new event with the panic level, using the global logger.
// A panic() call happens when the returned event is finalized with
// Msg/Msgf/Send.
func Panic() *zerolog.Event { return get().Panic() }
