package metrics

import (
	"context"
	"github.com/arandich/marketplace-sdk/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"time"
)

type OrderTimeMetric struct {
	OrderTime          prom.Histogram
	SendMessageTime    prom.Histogram
	ReceiveMessageTime prom.Histogram
	DeleteMessageTime  prom.Histogram
}

func NewOrderMetrics(cfg prometheus.Config) OrderTimeMetric {

	opts := prom.HistogramOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      "order_time",
		Help:      "Order time histogram",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 20},
	}

	opts2 := prom.HistogramOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      "send_message_time",
		Help:      "Send message time histogram",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 20},
	}

	opts3 := prom.HistogramOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      "receive_message_time",
		Help:      "Receive message time histogram",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 20},
	}

	opts4 := prom.HistogramOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      "delete_message_time",
		Help:      "Delete message time histogram",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 20},
	}
	return OrderTimeMetric{
		OrderTime:          prom.NewHistogram(opts),
		SendMessageTime:    prom.NewHistogram(opts2),
		ReceiveMessageTime: prom.NewHistogram(opts3),
		DeleteMessageTime:  prom.NewHistogram(opts4),
	}
}

func (m OrderTimeMetric) RecordOrderTime(ctx context.Context, orderTime time.Duration) {
	m.OrderTime.Observe(orderTime.Seconds())
}

func (m OrderTimeMetric) RecordSendMessageTime(ctx context.Context, sendMessageTime time.Duration) {
	m.SendMessageTime.Observe(sendMessageTime.Seconds())
}

func (m OrderTimeMetric) RecordReceiveMessageTime(ctx context.Context, receiveMessageTime time.Duration) {
	m.ReceiveMessageTime.Observe(receiveMessageTime.Seconds())
}

func (m OrderTimeMetric) RecordDeleteMessageTime(ctx context.Context, deleteMessageTime time.Duration) {
	m.DeleteMessageTime.Observe(deleteMessageTime.Seconds())
}
