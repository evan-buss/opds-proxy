package reqctx

import (
	"context"
	"log/slog"
	"testing"
)

func TestLoggerRoundTrip(t *testing.T) {
	base := context.Background()
	l := slog.Default()
	ctx := WithRequestLogger(base, l)
	got := Logger(ctx)
	if got != l {
		t.Fatalf("Logger did not round-trip")
	}
}

func TestIsLocalRoundTrip(t *testing.T) {
	base := context.Background()
	ctx := WithIsLocal(base, true)
	if !IsLocal(ctx) {
		t.Fatalf("expected IsLocal true")
	}
	ctx = WithIsLocal(base, false)
	if IsLocal(ctx) {
		t.Fatalf("expected IsLocal false")
	}
}
