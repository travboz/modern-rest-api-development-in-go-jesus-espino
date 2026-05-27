package tracer

import (
	"io"
)

// TODO: Add in tracing using Open Telemetry because book uses Jaeger but it's deprecated and useless now.
// See here: https://github.com/jaegertracing/jaeger-client-go

// Defines a tracer and initialises it.
func InitTracer(serviceName string) (io.Closer, error) {

	return nil, nil
}
