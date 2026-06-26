package infra

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Telemetry struct {
	traceProvider *trace.TracerProvider
	meterProvider *metric.MeterProvider
}

func InitTelemetry(ctx context.Context, serviceName string) (*Telemetry, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithResource(res),
	)
	
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	slog.InfoContext(ctx, "OpenTelemetry telemetry initialized")

	return &Telemetry{
		traceProvider: tp,
		meterProvider: mp,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) {
	if err := t.traceProvider.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "Erro ao desligar TracerProvider", "error", err)
	}
	if err := t.meterProvider.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "Erro ao desligar MeterProvider", "error", err)
	}
}
