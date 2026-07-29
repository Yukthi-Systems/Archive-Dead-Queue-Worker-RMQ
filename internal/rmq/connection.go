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

// Package rmq wires up a RabbitMQ connection and consumer (via
// github.com/wagslane/go-rabbitmq) around a pluggable Handler that decides
// what happens to each delivered message.
package rmq

import (
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

// Config holds the settings shared by NewConnection and NewConsumer: where
// to connect, which queue to consume, and how much work to do concurrently.
type Config struct {
	// URL is the AMQP connection string, e.g. "amqp://user:pass@host:5672/vhost".
	URL string
	// ConsumeQueue is the name of the queue to consume messages from.
	ConsumeQueue string
	// Prefetch is the per-consumer QoS prefetch count.
	Prefetch int
	// Concurrency is the number of goroutines processing deliveries concurrently.
	Concurrency int
}

// Connection wraps a single AMQP connection to a RabbitMQ broker.
type Connection struct {
	// Conn is the underlying go-rabbitmq connection.
	Conn *rabbitmq.Conn
}

// NewConnection dials the broker at cfg.URL and returns an open Connection.
// The caller is responsible for calling Close when done with it.
func NewConnection(cfg Config) (*Connection, error) {
	conn, err := rabbitmq.NewConn(
		cfg.URL,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, err
	}

	return &Connection{Conn: conn}, nil
}

// Close closes the underlying AMQP connection. Any error from closing is
// discarded, matching typical shutdown-path cleanup semantics.
func (c *Connection) Close() {
	_ = c.Conn.Close()
}
