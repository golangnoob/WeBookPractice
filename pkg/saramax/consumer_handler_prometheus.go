package saramax

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"

	"webooktrial/pkg/logger"
)

type HandlerOptions struct {
	Namespace  string
	Subsystem  string
	InstanceID string
}

type HandlerOption func(opts *HandlerOptions)

type HandlerWithPrometheus[T any] struct {
	consumerGroup string
	fn            func(msg *sarama.ConsumerMessage, t T) error
	l             logger.LoggerV1
	opts          *HandlerOptions
	// 监控消费耗时
	elapsedTimeSummary *prometheus.SummaryVec
	// 监控错误数
	errorCounter *prometheus.CounterVec
	// 监控消息积压
	offsetDelayGauge *prometheus.GaugeVec
}

func NewHandlerWithPrometheus[T any](consumerGroup string, fn func(msg *sarama.ConsumerMessage, t T) error,
	l logger.LoggerV1, opts ...HandlerOption) *HandlerWithPrometheus[T] {
	handler := &HandlerWithPrometheus[T]{
		consumerGroup: consumerGroup,
		fn:            fn,
		l:             l,
		opts:          &HandlerOptions{},
	}
	for _, opt := range opts {
		opt(handler.opts)
	}

	handler.elapsedTimeSummary = newElapsedTimeSummary(handler.opts)
	handler.errorCounter = newErrorCounter(handler.opts)
	handler.offsetDelayGauge = newOffsetDelayGauge(handler.opts)
	return handler
}

func newElapsedTimeSummary(opts *HandlerOptions) *prometheus.SummaryVec {
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      "elapsedTime",
		Help:      "记录消费耗时",
		ConstLabels: map[string]string{
			"instance_id": opts.InstanceID,
		},
		Objectives: map[float64]float64{
			0.9:  0.01,
			0.99: 0.001,
		},
	}, []string{"topic", "partition", "consumer_group"})
	prometheus.MustRegister(summary)
	return summary
}

func newOffsetDelayGauge(opts *HandlerOptions) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      "offset_delay",
		Help:      "监控已提交尚未被消费的信息数",
		ConstLabels: map[string]string{
			"instance_id": opts.InstanceID,
		},
	}, []string{"topic", "partition", "consumer_group"})
	prometheus.MustRegister(gauge)
	return gauge
}

func newErrorCounter(opts *HandlerOptions) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      "error_count",
		Help:      "记录错误数",
		ConstLabels: map[string]string{
			"instance_id": opts.InstanceID,
		},
	}, []string{"topic", "consumer_group", "error_type"})
	prometheus.MustRegister(counter)
	return counter
}

func (h *HandlerWithPrometheus[T]) Setup(session sarama.ConsumerGroupSession) error {
	h.l.Info("Setup DoNothing")
	return nil
}

func (h *HandlerWithPrometheus[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	h.l.Info("Cleanup DoNothing")
	return nil
}

func (h *HandlerWithPrometheus[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		highWaterMarkOffset := claim.HighWaterMarkOffset()
		offset := msg.Offset
		topic := msg.Topic
		partition := msg.Partition
		startTime := time.Now()

		h.recordOffsetDelay(topic, partition, highWaterMarkOffset-offset)

		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			h.l.Error("反序列化消息失败",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int64("partition", int64(msg.Partition)),
				logger.Int64("offset", msg.Offset))
			h.recordError(topic, partition, "反序列化消息失败")
			h.recordElapsedTime(topic, partition, startTime)
			continue
		}
		err = h.fn(msg, t)
		if err != nil {
			h.l.Error("反序列化消息失败",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int64("partition", int64(msg.Partition)),
				logger.Int64("offset", msg.Offset))
		}
		session.MarkMessage(msg, "")

		h.recordError(topic, partition, "业务处理失败")
		h.recordElapsedTime(topic, partition, startTime)
	}
	return nil
}

func (h *HandlerWithPrometheus[T]) recordElapsedTime(topic string, partition int32, startTime time.Time) {
	h.elapsedTimeSummary.WithLabelValues(topic, strconv.Itoa(int(partition)), h.consumerGroup).
		Observe(float64(time.Since(startTime).Milliseconds()))
}

func (h *HandlerWithPrometheus[T]) recordOffsetDelay(topic string, partition int32, unconsumedMsgs int64) {
	h.offsetDelayGauge.WithLabelValues(topic, strconv.Itoa(int(partition)), h.consumerGroup).Set(float64(unconsumedMsgs))
}

func (h *HandlerWithPrometheus[T]) recordError(topic string, partition int32, errType string) {
	h.errorCounter.WithLabelValues(topic, h.consumerGroup, errType).Inc()
}
