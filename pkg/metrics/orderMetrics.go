package metrics

import (
	"context"
	"github.com/arandich/marketplace-sdk/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"time"
)

type OrderTimeMetric struct {
	OrderTime prom.Histogram
}

func NewOrderMetrics(cfg prometheus.Config) OrderTimeMetric {

	opts := prom.HistogramOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      "order_time",
		Help:      "Order time histogram (ms)",
		Buckets:   prom.LinearBuckets(0.01, 0.05, 10),
	}

	return OrderTimeMetric{
		OrderTime: prom.NewHistogram(opts),
	}
}

func (m OrderTimeMetric) RecordOrderTime(ctx context.Context, orderTime time.Duration) {
	m.OrderTime.Observe(orderTime.Seconds() * 1000)
}
