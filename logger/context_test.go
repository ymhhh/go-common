package logger

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestTraceFields_nilContext(t *testing.T) {
	//nolint:staticcheck // TraceFields documents nil ctx as a valid empty input.
	if got := TraceFields(nil); got != nil {
		t.Fatalf("TraceFields(nil) = %v, want nil", got)
	}
}

func TestTraceFields_noSpan(t *testing.T) {
	if got := TraceFields(context.Background()); got != nil {
		t.Fatalf("TraceFields(background) = %v, want nil", got)
	}
}

func TestTraceFields_validSpan(t *testing.T) {
	traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	if err != nil {
		t.Fatalf("TraceIDFromHex: %v", err)
	}
	spanID, err := trace.SpanIDFromHex("0102030405060708")
	if err != nil {
		t.Fatalf("SpanIDFromHex: %v", err)
	}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	got := TraceFields(ctx)
	if got == nil {
		t.Fatal("TraceFields: nil")
	}
	if got["trace_id"] != traceID.String() {
		t.Fatalf("trace_id = %q, want %q", got["trace_id"], traceID.String())
	}
	if got["span_id"] != spanID.String() {
		t.Fatalf("span_id = %q, want %q", got["span_id"], spanID.String())
	}
}

func TestLFromContext_withSpan(t *testing.T) {
	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	entry := LFromContext(ctx)
	if entry.Data["trace_id"] != traceID.String() {
		t.Fatalf("trace_id = %q", entry.Data["trace_id"])
	}
	if entry.Data["span_id"] != spanID.String() {
		t.Fatalf("span_id = %q", entry.Data["span_id"])
	}
}

func TestLFromContext_withoutSpan(t *testing.T) {
	entry := LFromContext(context.Background())
	if _, ok := entry.Data["trace_id"]; ok {
		t.Fatalf("unexpected trace_id on background context: %v", entry.Data)
	}
}
