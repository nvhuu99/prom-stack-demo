package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
)

var logger *zap.Logger

const (
	OTEL_EXPORTER_OTLP_ENDPOINT = ""
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("cannot initialize zap logger: %v", err)
	}
	defer logger.Sync()

	tp, err := initTracer(ctx)
	if err != nil {
		logger.Fatal("failed to initialize tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(ctx) }()

	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(rootHandler), "root"))
	http.Handle("/hello", otelhttp.NewHandler(http.HandlerFunc(helloHandler), "hello"))

	srv := &http.Server{
		Addr: ":8080",
	}

	go func() {
		logger.Info("starting server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	traceID := trace.SpanFromContext(r.Context()).SpanContext().TraceID().String()
	logger.Info("root handler called", zap.String("trace_id", traceID))
	fmt.Fprintln(w, "Welcome to otel-golang-server")
}

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(OTEL_EXPORTER_OTLP_ENDPOINT), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("demo-web"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}


func helloHandler(w http.ResponseWriter, r *http.Request) {
	traceID := trace.SpanFromContext(r.Context()).SpanContext().TraceID().String()

	// Introduce random error with 20% probability
	if rand.Float64() < 0.2 {
		logger.Error("simulated error in hello handler", zap.String("trace_id", traceID))
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	logger.Info("hello handler called", zap.String("trace_id", traceID))
	fmt.Fprintln(w, "Hello, World!")
}

