package amqp

import (
	"context"
	"encoding/json"

	libamqp "gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"github.com/google/uuid"
)

type EventPublisher struct {
	producer libamqp.Producer
}

func NewEventPublisher(producer libamqp.Producer) *EventPublisher {
	return &EventPublisher{producer: producer}
}

type OrderCreatedEvent struct {
	OrderID     uuid.UUID `json:"order_id"`
	UserID      uuid.UUID `json:"user_id"`
	TotalAmount float64   `json:"total_amount"`
}

func (p *EventPublisher) PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.producer.Publish(ctx, libamqp.Delivery{
		RoutingKey:  "order.created",
		ContentType: "application/json",
		Type:        "order.created",
		Body:        body,
	})
}
