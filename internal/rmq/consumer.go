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

package rmq

import (
	"context"
	"time"

	rabbitmq "github.com/wagslane/go-rabbitmq"
)

// Consumer drives a durable queue subscription, dispatching each delivery
// to a Handler and acknowledging, rejecting, or requeueing it based on the
// Handler's returned Action.
type Consumer struct {
	consumer *rabbitmq.Consumer
	handler  Handler
	cfg      Config
}

// NewConsumer declares a durable consumer on cfg.ConsumeQueue over conn,
// configured with cfg.Prefetch QoS and cfg.Concurrency parallel workers.
// Deliveries are not consumed until Start is called.
func NewConsumer(
	conn *Connection,
	handler Handler,
	cfg Config,
) (*Consumer, error) {

	consumer, err := rabbitmq.NewConsumer(
		conn.Conn,
		cfg.ConsumeQueue,
		rabbitmq.WithConsumerOptionsQueueDurable,
		rabbitmq.WithConsumerOptionsQOSPrefetch(cfg.Prefetch),
		rabbitmq.WithConsumerOptionsConcurrency(cfg.Concurrency),
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		handler:  handler,
		cfg:      cfg,
	}, nil
}

// Start begins consuming deliveries and blocks until the consumer stops,
// either because ctx is cancelled or Close is called. Each delivery is
// processed with a derived context bounded to a 30-second timeout before
// being handed to the Handler.
func (c *Consumer) Start(ctx context.Context) error {

	return c.consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return c.handler.Process(processCtx, d)
	})
}

// Close stops the consumer, causing a blocked Start call to return.
func (c *Consumer) Close() {
	c.consumer.Close()
}
