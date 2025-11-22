package query

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type OrderQueryService interface {
	GetOrder(ctx context.Context, orderID uuid.UUID) (*Order, error)
}

type Order struct {
	OrderID    uuid.UUID
	UserID     uuid.UUID
	Status     string
	TotalPrice float64
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderItem struct {
	ProductID uuid.UUID
	Quantity  int
	Price     float64
}

func NewOrderQueryService(client mysql.ClientContext) OrderQueryService {
	return &orderQueryService{
		client: client,
	}
}

type orderQueryService struct {
	client mysql.ClientContext
}

func (s *orderQueryService) GetOrder(ctx context.Context, orderID uuid.UUID) (*Order, error) {
	orderData := struct {
		OrderID    uuid.UUID `db:"order_id"`
		UserID     uuid.UUID `db:"user_id"`
		Status     string    `db:"status"`
		TotalPrice float64   `db:"total_price"`
		CreatedAt  time.Time `db:"created_at"`
		UpdatedAt  time.Time `db:"updated_at"`
	}{}

	err := s.client.GetContext(
		ctx,
		&orderData,
		`SELECT order_id, user_id, status, total_price, created_at, updated_at FROM orders WHERE order_id = ?`,
		orderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	var itemsData []struct {
		ProductID uuid.UUID `db:"product_id"`
		Quantity  int       `db:"quantity"`
		Price     float64   `db:"price"`
	}

	err = s.client.SelectContext(
		ctx,
		&itemsData,
		`SELECT product_id, quantity, price FROM order_items WHERE order_id = ?`,
		orderID,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	items := make([]OrderItem, 0, len(itemsData))
	for _, item := range itemsData {
		items = append(items, OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return &Order{
		OrderID:    orderData.OrderID,
		UserID:     orderData.UserID,
		Status:     orderData.Status,
		TotalPrice: orderData.TotalPrice,
		Items:      items,
		CreatedAt:  orderData.CreatedAt,
		UpdatedAt:  orderData.UpdatedAt,
	}, nil
}
