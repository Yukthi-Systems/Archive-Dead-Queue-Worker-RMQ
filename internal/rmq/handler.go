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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Yukthi-Systems/Archive-Dead-Queue-Worker-RMQ/internal/logger"
	"github.com/Yukthi-Systems/Archive-Dead-Queue-Worker-RMQ/internal/model"
	"github.com/Yukthi-Systems/Archive-Dead-Queue-Worker-RMQ/internal/utils"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

// Handler processes a single delivery from a dead-letter queue and reports
// how the Consumer should acknowledge it back to the broker (e.g.
// rabbitmq.Ack, rabbitmq.NackRequeue). Implementations must be safe for
// concurrent use, since Consumer may invoke Process from multiple
// goroutines when Config.Concurrency > 1.
type Handler interface {
	Process(ctx context.Context, d rabbitmq.Delivery) rabbitmq.Action
}

// NewDefaultHandler returns a handler that archives dead-lettered messages
// under basePath, grouped into month folders, with a per-month metadata
// index describing every archived file.
func NewDefaultHandler(basePath string) Handler {
	return &defaultHandler{basePath: basePath}
}

// defaultHandler is the Handler returned by NewDefaultHandler. It archives
// every delivery to disk and never rejects a message outright: parse
// failures fall back to archiving the raw body, and only I/O errors trigger
// a requeue.
type defaultHandler struct {
	basePath string

	// guards concurrent appends to the same month's metadata file
	metaMu sync.Mutex
}

// Process archives one delivery under h.basePath/<YYYY-MM>/ and appends a
// matching entry to that month's metadata.jsonl. It attempts to parse the
// body as an EmailPayload and archive its EmailContent as a ".eml" file;
// if parsing fails (malformed JSON or a missing email_content field), it
// archives the raw message body instead as a ".raw.json" file.
//
// Process returns rabbitmq.Ack once the file and metadata entry are
// durably written, or rabbitmq.NackRequeue if a filesystem error prevents
// archiving, so the message is redelivered for a later retry.
func (h *defaultHandler) Process(ctx context.Context, d rabbitmq.Delivery) rabbitmq.Action {
	start := time.Now()
	reason := utils.DeadLetterReason(d.Headers)

	payload, extractErr := utils.ParsePayload(d.Body)

	archivedAt := time.Now()
	monthDir := filepath.Join(h.basePath, archivedAt.Format("2006-01"))
	if err := os.MkdirAll(monthDir, 0o755); err != nil {
		logger.Error().Err(err).Str("dir", monthDir).Msg("failed to create archive month directory")
		return rabbitmq.NackRequeue
	}

	messageID := payload.MessageID
	if messageID == "" {
		messageID = fmt.Sprintf("unknown-%d", archivedAt.UnixNano())
	}
	baseName := utils.SanitizeFileName(messageID)

	var content []byte
	isRaw := extractErr != nil
	var fileName string
	if isRaw {
		logger.Warn().Err(extractErr).Str("message_id", messageID).
			Msg("could not extract email_content from body; archiving raw message instead")
		fileName = baseName + ".raw.json"
		content = d.Body
	} else {
		fileName = baseName + ".eml"
		content = []byte(payload.EmailContent)
	}

	filePath := filepath.Join(monthDir, fileName)
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		logger.Error().Err(err).Str("file", filePath).Msg("failed to write archive file")
		return rabbitmq.NackRequeue
	}

	meta := model.ArchiveMeta{
		MessageID:    messageID,
		FileName:     fileName,
		ArchivedAt:   archivedAt,
		MessageTime:  payload.Timestamp,
		SizeBytes:    len(content),
		Reason:       reason,
		Exchange:     d.Exchange,
		RoutingKey:   d.RoutingKey,
		ContentIsRaw: isRaw,
	}
	if err := h.appendMeta(monthDir, meta); err != nil {
		logger.Error().Err(err).Msg("failed to write archive metadata")
		return rabbitmq.NackRequeue
	}

	logger.Info().
		Str("execution_time", time.Since(start).String()).
		Str("message_id", messageID).
		Str("file", filePath).
		Str("reason", reason).
		Bool("content_is_raw", isRaw).
		Msg("archived dead-lettered message")

	return rabbitmq.Ack
}

// appendMeta appends one JSON line to the month's metadata.jsonl index.
func (h *defaultHandler) appendMeta(monthDir string, meta model.ArchiveMeta) error {
	line, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	h.metaMu.Lock()
	defer h.metaMu.Unlock()

	f, err := os.OpenFile(filepath.Join(monthDir, "metadata.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(append(line, '\n'))
	return err
}
