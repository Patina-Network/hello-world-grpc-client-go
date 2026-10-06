package httpserve

import (
	"net/http"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type Metrics struct {
	registry *prometheus.Registry
	grpc     *grpcprom.ClientMetrics
}

func NewMetrics() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		grpc:     grpcprom.NewClientMetrics(grpcprom.WithClientHandlingTimeHistogram()),
	}

	m.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.grpc,
	)

	return m
}

func (m *Metrics) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return m.grpc.UnaryClientInterceptor()
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
