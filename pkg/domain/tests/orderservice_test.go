package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"order/pkg/domain/model"
	"order/pkg/domain/service"
)

func TestOrderService(t *testing.T) {
	repo := &mockOrderRepository{
		store: map[uuid.UUID]*model.Order{},
	}
	eventDispatcher := &mockEventDispatcher{}

	orderService := service.NewOrderService(repo, eventDispatcher)

	customerID := uuid.Must(uuid.NewV7())

	t.Run("Create order", func(t *testing.T) {
		eventDispatcher.events = []service.Event{}
		orderID, err := orderService.CreateOrder(customerID)
		require.NoError(t, err)

		require.NotNil(t, repo.store[orderID])
		require.Equal(t, model.Open, repo.store[orderID].Status)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.OrderCreated{}.Type(), eventDispatcher.events[0].Type())
	})

	t.Run("Add item to order", func(t *testing.T) {
		eventDispatcher.events = []service.Event{}
		orderID, _ := orderService.CreateOrder(customerID)
		eventDispatcher.events = []service.Event{}

		productID := uuid.Must(uuid.NewV7())
		itemID, err := orderService.AddItem(orderID, productID, 99.99)
		require.NoError(t, err)

		order := repo.store[orderID]
		require.Len(t, order.Items, 1)
		require.Equal(t, itemID, order.Items[0].ID)
		require.Equal(t, productID, order.Items[0].ProductID)
		require.Equal(t, 99.99, order.Items[0].Price)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.OrderItemChanged{}.Type(), eventDispatcher.events[0].Type())
	})

	t.Run("Delete item from order", func(t *testing.T) {
		eventDispatcher.events = []service.Event{} 
		orderID, _ := orderService.CreateOrder(customerID)
		productID := uuid.Must(uuid.NewV7())
		itemID, _ := orderService.AddItem(orderID, productID, 50.0)
		eventDispatcher.events = []service.Event{}

		err := orderService.DeleteItem(orderID, itemID)
		require.NoError(t, err)

		order := repo.store[orderID]
		require.Len(t, order.Items, 0)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.OrderItemChanged{}.Type(), eventDispatcher.events[0].Type())
	})

	t.Run("Cannot add item to non-open order", func(t *testing.T) {
		orderID, _ := orderService.CreateOrder(customerID)
		orderService.SetStatus(orderID, model.Pending)
		eventDispatcher.events = []service.Event{}

		productID := uuid.Must(uuid.NewV7())
		_, err := orderService.AddItem(orderID, productID, 100.0)
		require.Error(t, err)
		require.Equal(t, service.ErrInvalidOrderStatus, err)
	})

	t.Run("Change order status", func(t *testing.T) {
		eventDispatcher.events = []service.Event{} 
		orderID, _ := orderService.CreateOrder(customerID)
		eventDispatcher.events = []service.Event{} 

		err := orderService.SetStatus(orderID, model.Pending)
		require.NoError(t, err)

		order := repo.store[orderID]
		require.Equal(t, model.Pending, order.Status)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.OrderStatusChanged{}.Type(), eventDispatcher.events[0].Type())

		statusEvent := eventDispatcher.events[0].(model.OrderStatusChanged)
		require.Equal(t, model.Open, statusEvent.PreviousStatus)
		require.Equal(t, model.Pending, statusEvent.Status)
	})

	t.Run("Delete order", func(t *testing.T) {
		eventDispatcher.events = []service.Event{} 
		orderID, _ := orderService.CreateOrder(customerID)
		eventDispatcher.events = []service.Event{} 

		err := orderService.DeleteOrder(orderID)
		require.NoError(t, err)

		order := repo.store[orderID]
		require.NotNil(t, order.DeletedAt)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.OrderDeleted{}.Type(), eventDispatcher.events[0].Type())

		_, err = orderService.AddItem(orderID, uuid.Must(uuid.NewV7()), 100.0)
		require.Error(t, err)
		require.Equal(t, model.ErrOrderNotFound, err)
	})

	t.Run("Cannot delete item from non-open order", func(t *testing.T) {
		orderID, _ := orderService.CreateOrder(customerID)
		productID := uuid.Must(uuid.NewV7())
		itemID, _ := orderService.AddItem(orderID, productID, 50.0)
		orderService.SetStatus(orderID, model.Paid)

		err := orderService.DeleteItem(orderID, itemID)
		require.Error(t, err)
		require.Equal(t, service.ErrInvalidOrderStatus, err)
	})

	t.Run("Set same status does not dispatch event", func(t *testing.T) {
		eventDispatcher.events = []service.Event{} 
		orderID, _ := orderService.CreateOrder(customerID)
		eventDispatcher.events = []service.Event{} 

		err := orderService.SetStatus(orderID, model.Open)
		require.NoError(t, err)
		require.Len(t, eventDispatcher.events, 0)
	})
}

var _ model.OrderRepository = &mockOrderRepository{}

type mockOrderRepository struct {
	store map[uuid.UUID]*model.Order
}

func (m mockOrderRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m mockOrderRepository) Store(order *model.Order) error {
	m.store[order.ID] = order
	return nil
}

func (m mockOrderRepository) Find(id uuid.UUID) (*model.Order, error) {
	if order, ok := m.store[id]; ok && order.DeletedAt == nil {
		return order, nil
	}
	return nil, model.ErrOrderNotFound
}

func (m mockOrderRepository) Delete(id uuid.UUID) error {
	if order, ok := m.store[id]; ok && order.DeletedAt == nil {
		order.DeletedAt = toPtr(time.Now())
		return nil
	}
	return model.ErrOrderNotFound
}

var _ service.EventDispatcher = &mockEventDispatcher{}

type mockEventDispatcher struct {
	events []service.Event
}

func (m *mockEventDispatcher) Dispatch(event service.Event) error {
	m.events = append(m.events, event)
	return nil
}

func toPtr[V any](v V) *V {
	return &v
}
