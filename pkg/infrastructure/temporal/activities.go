package temporal

import (
	"context"

	appmodel "order/pkg/application/model"
	"order/pkg/application/service"
	domainservice "order/pkg/domain/service"

	"github.com/google/uuid"
)

type Activities struct {
	UoW                 service.UnitOfWork
	ProductService      service.ProductService
	PaymentService      service.PaymentService
	NotificationService service.NotificationService
}

func NewActivities(
	uow service.UnitOfWork,
	productService service.ProductService,
	paymentService service.PaymentService,
	notificationService service.NotificationService,
) *Activities {
	return &Activities{
		UoW:                 uow,
		ProductService:      productService,
		PaymentService:      paymentService,
		NotificationService: notificationService,
	}
}

func (a *Activities) CreateOrderActivity(ctx context.Context, order appmodel.Order) (uuid.UUID, error) {
	var orderID uuid.UUID
	err := a.UoW.Execute(ctx, func(provider service.RepositoryProvider) error {
		domainService := domainservice.NewOrderService(provider.OrderRepository(ctx), &service.NoOpEventDispatcher{})
		var err error
		orderID, err = domainService.CreateOrder(order.UserID)
		return err
	})
	return orderID, err
}

func (a *Activities) GetProductPriceActivity(ctx context.Context, productID uuid.UUID) (float64, error) {
	return a.ProductService.GetPrice(ctx, productID)
}

func (a *Activities) AddItemActivity(ctx context.Context, orderID uuid.UUID, productID uuid.UUID, price float64, quantity int) error {
	return a.UoW.Execute(ctx, func(provider service.RepositoryProvider) error {
		domainService := domainservice.NewOrderService(provider.OrderRepository(ctx), &service.NoOpEventDispatcher{})
		for i := 0; i < quantity; i++ {
			_, err := domainService.AddItem(orderID, productID, price)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (a *Activities) ProcessPaymentActivity(ctx context.Context, userID, orderID uuid.UUID, amount float64) error {
	return a.PaymentService.ProcessPayment(ctx, userID, orderID, amount)
}

func (a *Activities) SendNotificationActivity(ctx context.Context, userID uuid.UUID, message string) error {
	return a.NotificationService.SendNotification(ctx, userID, message)
}
