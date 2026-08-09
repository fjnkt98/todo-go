package server

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

func SetupOTel(ctx context.Context) (func(context.Context) error, error) {
	var shutdowns []func(context.Context) error
	var err error

	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdowns {
			err = errors.Join(err, fn(ctx))
		}
		shutdowns = nil
		return err
	}

	// propagator
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)

	// tracer provider
	traceExporter, err := stdouttrace.New()
	if err != nil {
		err = errors.Join(err, shutdown(ctx))
		return shutdown, err
	}
	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(traceExporter, trace.WithBatchTimeout(time.Second)))
	if err != nil {
		err = errors.Join(err, shutdown(ctx))
		return shutdown, err
	}
	shutdowns = append(shutdowns, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// meter provider
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		err = errors.Join(err, shutdown(ctx))
		return shutdown, err
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(3*time.Second))))
	shutdowns = append(shutdowns, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// logger provider
	logExporter, err := stdoutlog.New()
	if err != nil {
		err = errors.Join(err, shutdown(ctx))
		return shutdown, err
	}
	loggerProvider := log.NewLoggerProvider(log.WithProcessor(log.NewBatchProcessor(logExporter)))
	shutdowns = append(shutdowns, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, nil
}
