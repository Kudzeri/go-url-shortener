package random

import (
	"strings"
	"testing"
)

func TestNewRandomString(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{name: "size = 1", size: 1},
		{name: "size = 5", size: 5},
		{name: "size = 10", size: 10},
		{name: "size = 20", size: 20},
		{name: "size = 32", size: 32},
		{name: "zero size", size: 0, wantErr: true},
		{name: "negative size", size: -1, wantErr: true},
	}

	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewRandomString(tt.size)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				if result != "" {
					t.Errorf("result = %q, want empty string", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != tt.size {
				t.Errorf("length = %d, want %d", len(result), tt.size)
			}

			for _, char := range result {
				if !strings.ContainsRune(alphabet, char) {
					t.Errorf("unexpected character: %q", char)
				}
			}
		})
	}
}
