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

// republish reads a RabbitMQ management API "get message(s)" JSON export
// (e.g. the response saved from POST /api/queues/<vhost>/<queue>/get)
// and publishes one of its messages, verbatim, into a queue on another
// broker. Handy for taking a message peeked off the dead_queue and
// replaying it against a test RabbitMQ instance.
package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

// mgmtMessage is one entry of the RabbitMQ management API's "get
// message(s)" JSON response.
type mgmtMessage struct {
	// Payload is the message body, encoded as described by PayloadEncoding.
	Payload string `json:"payload"`
	// PayloadEncoding is either "string" (Payload is used as-is) or
	// "base64" (Payload must be base64-decoded to recover the body).
	PayloadEncoding string `json:"payload_encoding"`
}

// main parses the -file/-index/-url/-queue flags, loads the selected
// message from the management API export, decodes its payload, and
// publishes it verbatim to the target queue via the default exchange.
func main() {
	file := flag.String("file", "message.txt", "path to the management API 'get message' JSON export")
	index := flag.Int("index", 0, "index of the message in the export array to republish")
	url := flag.String("url", os.Getenv("TEST_AMQP_URL"), "amqp url of the target (test) broker, e.g. amqp://user:pass@host:5672/")
	queue := flag.String("queue", "", "target queue name to publish into (delivered via the default exchange)")
	flag.Parse()

	if *url == "" || *queue == "" {
		fmt.Fprintln(os.Stderr, "usage: republish -file message.txt -url amqp://user:pass@host:5672/ -queue test_queue")
		os.Exit(1)
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read file:", err)
		os.Exit(1)
	}

	var messages []mgmtMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		fmt.Fprintln(os.Stderr, "parse json:", err)
		os.Exit(1)
	}
	if *index >= len(messages) {
		fmt.Fprintf(os.Stderr, "index %d out of range (export has %d message(s))\n", *index, len(messages))
		os.Exit(1)
	}
	msg := messages[*index]

	var body []byte
	if msg.PayloadEncoding == "base64" {
		body, err = base64.StdEncoding.DecodeString(msg.Payload)
		if err != nil {
			fmt.Fprintln(os.Stderr, "decode base64 payload:", err)
			os.Exit(1)
		}
	} else {
		body = []byte(msg.Payload)
	}

	conn, err := amqp.Dial(*url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial:", err)
		os.Exit(1)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "open channel:", err)
		os.Exit(1)
	}
	defer ch.Close()

	// publish via the default exchange ("") -> routing key == queue name
	// delivers straight into that queue.
	if err := ch.Publish("", *queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "publish:", err)
		os.Exit(1)
	}

	fmt.Printf("published %d bytes to queue %q on %s\n", len(body), *queue, *url)
}
