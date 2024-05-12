package metrics

import (
	sdkPrometheus "github.com/arandich/marketplace-sdk/prometheus"
)

type Metrics struct {
	// Base metrics.
	BaseMetrics sdkPrometheus.Metrics
	// OrderTimeMetric
	OrderTimeMetric OrderTimeMetric
}

func New(baseMetrics sdkPrometheus.Metrics, orderTimeMetric OrderTimeMetric) Metrics {
	return Metrics{
		BaseMetrics:     baseMetrics,
		OrderTimeMetric: orderTimeMetric,
	}
}
