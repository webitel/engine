package app

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	trace.Tracer
}

func NewTrace() *Tracer {
	tp := otel.GetTracerProvider()

	return &Tracer{
		Tracer: tp.Tracer("engine"),
	}
}

func (a *App) Tracer() *Tracer {
	return a.tracer
}
