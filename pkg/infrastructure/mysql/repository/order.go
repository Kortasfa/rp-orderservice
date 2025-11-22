package repository

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"order/pkg/domain/model"
)

func NewOrderRepository(ctx context.Context, client mysql.ClientContext) model.OrderRepository {
	return &orderRepository{
		ctx:    ctx,
		client: client,
	}
}

type orderRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *orderRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (r *orderRepository) Store(order *model.Order) error {
	_, err := r.client.ExecContext(r.ctx,
		`
INSERT INTO orders (order_id, user_id, status, total_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
status = VALUES(status),
updated_at = VALUES(updated_at)
`,
		order.ID,
		order.CustomerID,
		order.Status,
		0.0, // Total price logic to be refined
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	// Simple implementation: delete all items and re-insert (not efficient but simple for now)
	_, err = r.client.ExecContext(r.ctx, `DELETE FROM order_items WHERE order_id = ?`, order.ID)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, item := range order.Items {
		_, err = r.client.ExecContext(r.ctx,
			`INSERT INTO order_items (item_id, order_id, product_id, quantity, price) VALUES (?, ?, ?, ?, ?)`,
			item.ID,
			order.ID,
			item.ProductID,
			1, // Quantity default 1 for now
			item.Price,
		)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *orderRepository) Find(id uuid.UUID) (*model.Order, error) {
	orderData := struct {
		OrderID    uuid.UUID `db:"order_id"`
		UserID     uuid.UUID `db:"user_id"`
		Status     int       `db:"status"`
		TotalPrice float64   `db:"total_price"`
		CreatedAt  time.Time `db:"created_at"`
		UpdatedAt  time.Time `db:"updated_at"`
	}{}

	err := r.client.GetContext(
		r.ctx,
		&orderData,
		`SELECT order_id, user_id, status, total_price, created_at, updated_at FROM orders WHERE order_id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrOrderNotFound)
		}
		return nil, errors.WithStack(err)
	}

	var itemsData []struct {
		ItemID    uuid.UUID `db:"item_id"`
		ProductID uuid.UUID `db:"product_id"`
		Price     float64   `db:"price"`
	}

	err = r.client.SelectContext(
		r.ctx,
		&itemsData,
		`SELECT item_id, product_id, price FROM order_items WHERE order_id = ?`,
		id,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	items := make([]model.Item, 0, len(itemsData))
	for _, item := range itemsData {
		items = append(items, model.Item{
			ID:        item.ItemID,
			ProductID: item.ProductID,
			Price:     item.Price,
		})
	}

	return &model.Order{
		ID:         orderData.OrderID,
		CustomerID: orderData.UserID,
		Status:     model.OrderStatus(orderData.Status),
		Items:      items,
		CreatedAt:  orderData.CreatedAt,
		UpdatedAt:  orderData.UpdatedAt,
	}, nil
}

func (r *orderRepository) Delete(id uuid.UUID) error {
	// Soft delete not implemented in DB schema yet, so hard delete
	_, err := r.client.ExecContext(r.ctx, `DELETE FROM orders WHERE order_id = ?`, id)
	return errors.WithStack(err)
}
