package logger

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// TraceFields returns trace_id and span_id when ctx carries a valid OTel span.
// Returns nil when ctx is nil or the span context is invalid.
func TraceFields(ctx context.Context) Fields {
	if ctx == nil {
		return nil
	}
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return nil
	}
	return Fields{
		"trace_id": sc.TraceID().String(),
		"span_id":  sc.SpanID().String(),
	}
}

// LFromContext returns the global logger entry with trace_id and span_id fields
// when ctx carries a valid OpenTelemetry span (e.g. after otelgin / otelgrpc middleware).
// Falls back to L() when no valid span is present.
func LFromContext(ctx context.Context) *logrus.Entry {
	entry := L()
	if fields := TraceFields(ctx); len(fields) > 0 {
		entry = entry.WithFields(fields)
	}
	return entry
}
