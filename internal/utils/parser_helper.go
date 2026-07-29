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

// Package utils provides small, focused helpers used by the rmq handler:
// parsing a dead-lettered message body, extracting its dead-letter reason,
// and sanitizing values for use as filenames.
package utils

import (
	"encoding/json"
	"errors"

	"github.com/Yukthi-Systems/Archive-Dead-Queue-Worker-RMQ/internal/model"
)

// ParsePayload extracts the fields we need from the dead-lettered body. The
// body must be valid JSON with a non-empty email_content field; anything
// else (malformed JSON, missing field) is treated as unparseable and the
// caller archives the raw body instead.
func ParsePayload(body []byte) (model.EmailPayload, error) {
	var p model.EmailPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return p, err
	}
	if p.EmailContent == "" {
		return p, errors.New("email_content field not found in message body")
	}
	return p, nil
}
