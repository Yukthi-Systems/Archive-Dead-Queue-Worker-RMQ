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

// Package model defines the data shapes shared across the archiver: the
// per-message metadata record written to each month's index, and the
// expected shape of a dead-lettered message body.
package model

import "time"

// ArchiveMeta is one line of a month's metadata.jsonl index, describing a
// single archived message and where its content was written to disk.
type ArchiveMeta struct {
	// MessageID is the dead-lettered message's ID, or a generated
	// "unknown-<unixnano>" placeholder if none was present.
	MessageID string `json:"message_id"`
	// FileName is the archived file's name within the month directory.
	FileName string `json:"file_name"`
	// ArchivedAt is when the message was archived.
	ArchivedAt time.Time `json:"archived_at"`
	// MessageTime is the timestamp reported inside the message body, if any.
	MessageTime string `json:"message_timestamp,omitempty"`
	// SizeBytes is the size in bytes of the archived file's content.
	SizeBytes int `json:"size_bytes"`
	// Reason is the dead-letter reason extracted from the AMQP headers.
	Reason string `json:"reason"`
	// Exchange is the source exchange the message was published to.
	Exchange string `json:"exchange,omitempty"`
	// RoutingKey is the routing key the message was published with.
	RoutingKey string `json:"routing_key,omitempty"`
	// ContentIsRaw is true if the body could not be parsed as EmailPayload
	// and the raw message body was archived instead of EmailContent.
	ContentIsRaw bool `json:"content_is_raw"`
}

// EmailPayload models the fields the archiver cares about in a
// dead-lettered message body. Fields beyond these are ignored during
// unmarshaling.
type EmailPayload struct {
	// Timestamp is the message's self-reported time, copied verbatim into
	// ArchiveMeta.MessageTime.
	Timestamp string `json:"timestamp"`
	// MessageID uniquely identifies the message.
	MessageID string `json:"message_id"`
	// EmailContent is the raw email content to archive; it must be
	// non-empty for the payload to be considered successfully parsed (see
	// utils.ParsePayload).
	EmailContent string `json:"email_content"`
}
