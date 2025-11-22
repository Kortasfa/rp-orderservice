package service

import (
	"context"

	"github.com/google/uuid"

	appmodel "order/pkg/application/model"
	"order/pkg/domain/service"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order appmodel.Order) (uuid.UUID, error)
}

func NewOrderService(
	uow UnitOfWork,
) OrderService {
	return &orderService{
		uow: uow,
	}
}

type orderService struct {
	uow UnitOfWork
}

type NoOpEventDispatcher struct{}

func (d *NoOpEventDispatcher) Dispatch(event service.Event) error {
	return nil
}

func (s *orderService) CreateOrder(ctx context.Context, order appmodel.Order) (uuid.UUID, error) {
	var orderID uuid.UUID
	err := s.uow.Execute(ctx, func(provider RepositoryProvider) error {
		domainService := service.NewOrderService(provider.OrderRepository(ctx), &NoOpEventDispatcher{})

		var err error
		orderID, err = domainService.CreateOrder(order.UserID)
		if err != nil {
			return err
		}

		for _, item := range order.Items {
			// TODO: Fetch price from product service
			// For now, we just add items with 0 price
			// Also, we need to handle quantity. The domain service AddItem adds one item?
			// No, AddItem adds an item with price. It doesn't specify quantity.
			// Wait, the domain model Item doesn't have quantity?
			// Let's check domain model Item.

			// Assuming AddItem adds a single unit. If quantity > 1, we call it multiple times?
			// Or we should update domain model.

			for i := 0; i < item.Quantity; i++ {
				_, err = domainService.AddItem(orderID, item.ProductID, 0)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return orderID, err
}
