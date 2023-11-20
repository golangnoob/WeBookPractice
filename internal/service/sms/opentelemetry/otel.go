package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"webooktrial/internal/service/sms"
)

type Service struct {
	svc    sms.Service
	tracer trace.Tracer
}

func NewService(svc sms.Service) *Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("webooktrial/internal/service/sms/opentelemetry")
	return &Service{
		svc:    svc,
		tracer: tracer,
	}
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	//tracer := s.tracerProvider.Tracer()
	ctx, span := s.tracer.Start(ctx, "sms_send"+biz, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End(trace.WithStackTrace(true))
	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
