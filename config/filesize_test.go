package config

import (
	"testing"
)

func TestFileSize(t *testing.T) {
	fs := FileSize(33554432)
	if got := fs.String(); got != "32.0MB" {
		t.Errorf("Expected 32M, got %s", got)
	}
}
