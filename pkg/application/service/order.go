package service

import (
	"context"

	"github.com/google/uuid"

	appmodel "order/pkg/application/model"
	"order/pkg/domain/service"
	infraamqp "order/pkg/infrastructure/amqp"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order appmodel.Order) (uuid.UUID, error)
	CreateOrderAsync(ctx context.Context, order appmodel.Order) (uuid.UUID, error)
}

type ProductService interface {
	GetPrice(ctx context.Context, productID uuid.UUID) (float64, error)
}

type PaymentService interface {
	ProcessPayment(ctx context.Context, userID, orderID uuid.UUID, amount float64) error
}

type NotificationService interface {
	SendNotification(ctx context.Context, userID uuid.UUID, message string) error
}

type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, event infraamqp.OrderCreatedEvent) error
}

type WorkflowStarter interface {
	StartCreateOrderWorkflow(ctx context.Context, order appmodel.Order) (uuid.UUID, error)
}

func NewOrderService(
	uow UnitOfWork,
	productService ProductService,
	eventPublisher EventPublisher,
	workflowStarter WorkflowStarter,
) OrderService {
	return &orderService{
		uow:             uow,
		productService:  productService,
		eventPublisher:  eventPublisher,
		workflowStarter: workflowStarter,
	}
}

type orderService struct {
	uow             UnitOfWork
	productService  ProductService
	eventPublisher  EventPublisher
	workflowStarter WorkflowStarter
}

type NoOpEventDispatcher struct{}

func (d *NoOpEventDispatcher) Dispatch(event service.Event) error {
	return nil
}

func (s *orderService) CreateOrder(ctx context.Context, order appmodel.Order) (uuid.UUID, error) {
	return s.workflowStarter.StartCreateOrderWorkflow(ctx, order)
}

func (s *orderService) CreateOrderAsync(ctx context.Context, order appmodel.Order) (uuid.UUID, error) {
	var orderID uuid.UUID
	err := s.uow.Execute(ctx, func(provider RepositoryProvider) error {
		domainService := service.NewOrderService(provider.OrderRepository(ctx), &NoOpEventDispatcher{})

		var err error
		orderID, err = domainService.CreateOrder(order.UserID)
		if err != nil {
			return err
		}

		var totalAmount float64

		for _, item := range order.Items {
			price, err := s.productService.GetPrice(ctx, item.ProductID)
			if err != nil {
				return err
			}

			totalAmount += price * float64(item.Quantity)

			for i := 0; i < item.Quantity; i++ {
				_, err = domainService.AddItem(orderID, item.ProductID, price)
				if err != nil {
					return err
				}
			}
		}

		event := infraamqp.OrderCreatedEvent{
			OrderID:     orderID,
			UserID:      order.UserID,
			TotalAmount: totalAmount,
		}
		if err := s.eventPublisher.PublishOrderCreated(ctx, event); err != nil {
			return err
		}

		return nil
	})
	return orderID, err
}
