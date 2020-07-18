package main

import (
	"strings"
)

func stripExeSuffix(p string) string {
	if strings.HasSuffix(strings.ToLower(p), ".exe") ||
		strings.HasSuffix(strings.ToLower(p), ".bin") {
		p = p[:len(p)-4]
	}
	return p
}
