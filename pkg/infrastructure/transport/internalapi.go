package transport

import (
	"context"

	"github.com/google/uuid"

	"order/api/server/orderinternal"
	appmodel "order/pkg/application/model"
	"order/pkg/application/service"
	"order/pkg/infrastructure/mysql/query"
)

func NewOrderInternalAPI(
	orderQueryService query.OrderQueryService,
	orderService service.OrderService,
) orderinternal.OrderInternalServiceServer {
	return &orderInternalAPI{
		orderQueryService: orderQueryService,
		orderService:      orderService,
	}
}

type orderInternalAPI struct {
	orderQueryService query.OrderQueryService
	orderService      service.OrderService

	orderinternal.UnimplementedOrderInternalServiceServer
}

func (a *orderInternalAPI) CreateOrder(ctx context.Context, request *orderinternal.CreateOrderRequest) (*orderinternal.CreateOrderResponse, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, err
	}

	var items []appmodel.OrderItem
	for _, item := range request.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, err
		}
		items = append(items, appmodel.OrderItem{
			ProductID: productID,
			Quantity:  int(item.Quantity),
		})
	}

	orderID, err := a.orderService.CreateOrder(ctx, appmodel.Order{
		UserID: userID,
		Items:  items,
	})
	if err != nil {
		return nil, err
	}

	return &orderinternal.CreateOrderResponse{
		OrderID: orderID.String(),
	}, nil
}

func (a *orderInternalAPI) GetOrder(ctx context.Context, request *orderinternal.GetOrderRequest) (*orderinternal.GetOrderResponse, error) {
	orderID, err := uuid.Parse(request.OrderID)
	if err != nil {
		return nil, err
	}

	order, err := a.orderQueryService.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return &orderinternal.GetOrderResponse{}, nil
	}

	var items []*orderinternal.OrderItem
	for _, item := range order.Items {
		items = append(items, &orderinternal.OrderItem{
			ProductID: item.ProductID.String(),
			Quantity:  int32(item.Quantity),
		})
	}

	return &orderinternal.GetOrderResponse{
		Order: &orderinternal.Order{
			OrderID:    order.OrderID.String(),
			UserID:     order.UserID.String(),
			Status:     order.Status,
			Items:      items,
			TotalPrice: order.TotalPrice,
		},
	}, nil
}

func (a *orderInternalAPI) CreateOrderAsync(ctx context.Context, request *orderinternal.CreateOrderRequest) (*orderinternal.CreateOrderResponse, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, err
	}

	var items []appmodel.OrderItem
	for _, item := range request.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, err
		}
		items = append(items, appmodel.OrderItem{
			ProductID: productID,
			Quantity:  int(item.Quantity),
		})
	}

	orderID, err := a.orderService.CreateOrderAsync(ctx, appmodel.Order{
		UserID: userID,
		Items:  items,
	})
	if err != nil {
		return nil, err
	}

	return &orderinternal.CreateOrderResponse{
		OrderID: orderID.String(),
	}, nil
}
