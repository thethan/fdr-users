package test_helpers

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/global"
)

func TestingTracer() otel.Tracer {
	tracer := global.Tracer("testing-tracer")
	return tracer
}
