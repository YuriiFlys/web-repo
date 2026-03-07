package httpx

import (
	"net/http"
	"testing"
)

func TestErr(t *testing.T) {
	err := Err(CodeBadRequest, "bad")
	if err.Code != CodeBadRequest || err.Message != "bad" {
		t.Fatalf("err = %+v", err)
	}
}

func TestStatusFor(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{CodeNotFound, http.StatusNotFound},
		{CodeUnauthorized, http.StatusUnauthorized},
		{"OTHER", http.StatusBadRequest},
	}
	for _, tt := range tests {
		if got := StatusFor(tt.code); got != tt.want {
			t.Fatalf("StatusFor(%q) = %d, want %d", tt.code, got, tt.want)
		}
	}
}
