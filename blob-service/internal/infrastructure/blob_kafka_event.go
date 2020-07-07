package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
	"gocloud.dev/pubsub"
	"strings"
	"sync"
	"time"
)

type BlobKafkaEvent struct {
	cfg *config.Kernel
	mu  *sync.Mutex
}

func NewBlobKafkaEvent(cfg *config.Kernel) *BlobKafkaEvent {
	return &BlobKafkaEvent{
		cfg: cfg,
		mu:  new(sync.Mutex),
	}
}

func (e BlobKafkaEvent) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "blob_kafka_" + action,
		MaxRequests: 1,
		Interval:    0,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: nil,
	}

	return gobreaker.NewCircuitBreaker(st)
}

func (e *BlobKafkaEvent) Uploaded(ctx context.Context, blob domain.Blob, snapshot *domain.Blob) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	urlPool := []string{blob.Url}
	urlJSON, err := json.Marshal(&urlPool)
	if err != nil {
		return err
	}

	snapshotJSON := []byte("")
	if snapshot != nil {
		snapshotJSON, err = json.Marshal(&snapshot)
		if err != nil {
			return err
		}
	}

	parentSpan := trace.FromContext(ctx)
	defer parentSpan.End()
	ctxT, span := trace.StartSpanWithRemoteParent(ctx, "blob: upload", parentSpan.SpanContext())
	defer span.End()

	p, err := eventbus.NewKafkaProducer(ctxT,
		fmt.Sprintf("%s_%s", strings.ToUpper(blob.Service), domain.BlobUploaded))
	if err != nil {
		return err
	}
	defer p.Shutdown(ctxT)

	transaction := eventbus.Transaction{
		ID:        uuid.New().String(),
		RootID:    blob.ID,
		SpanID:    span.SpanContext().SpanID.String(),
		TraceID:   span.SpanContext().TraceID.String(),
		Operation: domain.BlobUploaded,
		Snapshot:  string(snapshotJSON),
	}
	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, urlJSON)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id": transaction.ID,
			"root_id":        transaction.RootID,
			"span_id":        transaction.SpanID,
			"trace_id":       transaction.TraceID,
			"operation":      transaction.Operation,
			"snapshot":       transaction.Snapshot,
			"service":        event.ServiceName,
			"event_id":       event.ID,
			"event_type":     event.EventType,
			"priority":       event.Priority,
			"provider":       event.Provider,
			"dispatch_time":  event.DispatchTime,
		},
		BeforeSend: nil,
	}

	_, err = e.defaultCircuitBreaker("uploaded").Execute(func() (interface{}, error) {
		return nil, p.Send(ctxT, m)
	})
	return err
}

func (e *BlobKafkaEvent) Removed(ctx context.Context, rootID, service string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rootPool := []string{rootID}
	rootJSON, err := json.Marshal(&rootPool)
	if err != nil {
		return err
	}

	parentSpan := trace.FromContext(ctx)
	defer parentSpan.End()

	ctxT, span := trace.StartSpanWithRemoteParent(ctx, "blob: removed", parentSpan.SpanContext())
	defer span.End()

	p, err := eventbus.NewKafkaProducer(ctxT,
		fmt.Sprintf("%s_%s", strings.ToUpper(service), domain.BlobRemoved))
	if err != nil {
		return err
	}
	defer p.Shutdown(ctxT)

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, rootJSON)
	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"service":       event.ServiceName,
			"event_id":      event.ID,
			"event_type":    event.EventType,
			"priority":      event.Priority,
			"provider":      event.Provider,
			"dispatch_time": event.DispatchTime,
		},
		BeforeSend: nil,
	}

	_, err = e.defaultCircuitBreaker("removed").Execute(func() (interface{}, error) {
		return nil, p.Send(ctxT, m)
	})
	return err
}
