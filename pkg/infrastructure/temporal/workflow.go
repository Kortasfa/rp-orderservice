package temporal

import (
	"time"

	appmodel "order/pkg/application/model"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue               = "ORDER_TASK_QUEUE"
	CreateOrderWorkflowName = "CreateOrderWorkflow"
)

type CreateOrderWorkflowInput struct {
	Order appmodel.Order
}

type CreateOrderWorkflowResult struct {
	OrderID uuid.UUID
}

func CreateOrderWorkflow(ctx workflow.Context, input CreateOrderWorkflowInput) (CreateOrderWorkflowResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var orderID uuid.UUID
	// 1. Create Order in DB (Pending)
	err := workflow.ExecuteActivity(ctx, "CreateOrderActivity", input.Order).Get(ctx, &orderID)
	if err != nil {
		return CreateOrderWorkflowResult{}, err
	}

	// 2. Process Items (Get Price and Add to Order)
	var totalAmount float64
	for _, item := range input.Order.Items {
		var price float64
		err = workflow.ExecuteActivity(ctx, "GetProductPriceActivity", item.ProductID).Get(ctx, &price)
		if err != nil {
			return CreateOrderWorkflowResult{}, err
		}
		totalAmount += price * float64(item.Quantity)

		err = workflow.ExecuteActivity(ctx, "AddItemActivity", orderID, item.ProductID, price, item.Quantity).Get(ctx, nil)
		if err != nil {
			return CreateOrderWorkflowResult{}, err
		}
	}

	// 3. Process Payment
	err = workflow.ExecuteActivity(ctx, "ProcessPaymentActivity", input.Order.UserID, orderID, totalAmount).Get(ctx, nil)
	if err != nil {
		return CreateOrderWorkflowResult{}, err
	}

	// 4. Send Notification
	_ = workflow.ExecuteActivity(ctx, "SendNotificationActivity", input.Order.UserID, "Order created via Temporal").Get(ctx, nil)

	return CreateOrderWorkflowResult{OrderID: orderID}, nil
}
