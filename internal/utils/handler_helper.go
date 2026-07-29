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

package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var unsafeFileNameChars = regexp.MustCompile(`[^A-Za-z0-9._-]`)

// SanitizeFileName keeps message IDs safe to use as a filename component.
func SanitizeFileName(s string) string {
	s = unsafeFileNameChars.ReplaceAllString(s, "_")
	s = strings.Trim(s, "._")
	if s == "" {
		return "unknown"
	}
	return s
}

// DeadLetterReason pulls a human-readable failure reason out of the AMQP
// headers, preferring an explicit x-dead-letter-reason header and falling
// back to the standard RabbitMQ x-death entry.
func DeadLetterReason(headers map[string]interface{}) string {
	if v, ok := headers["x-dead-letter-reason"]; ok {
		return fmt.Sprintf("%v", v)
	}

	if v, ok := headers["x-death"]; ok {
		if deaths, ok := v.([]interface{}); ok && len(deaths) > 0 {
			if first, ok := deaths[0].(map[string]interface{}); ok {
				if r, ok := first["reason"]; ok {
					return fmt.Sprintf("%v", r)
				}
			}
		}
	}

	return "unknown"
}
