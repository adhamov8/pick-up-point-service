package tracer

import (
	"context"
	"log"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	traceconfig "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

func SetupTracer(ctx context.Context, serviceName string) {
	cfg := traceconfig.Configuration{
		ServiceName: serviceName,
		Sampler: &traceconfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &traceconfig.ReporterConfig{
			LogSpans: true,
		},
	}

	tracer, closer, err := cfg.NewTracer(
		traceconfig.Logger(jaeger.StdLogger),
		traceconfig.Metrics(prometheus.New()),
	)
	if err != nil {
		log.Printf("Error initializing Jaeger tracer: %v", err)
		return
	}

	opentracing.SetGlobalTracer(tracer)

	var once sync.Once
	go func() {
		<-ctx.Done()
		once.Do(func() {
			if err := closer.Close(); err != nil {
				log.Printf("Error closing tracer: %v", err)
			}
		})
	}()
}
