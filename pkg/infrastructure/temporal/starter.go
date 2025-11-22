package temporal

import (
	"context"

	appmodel "order/pkg/application/model"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type WorkflowStarterImpl struct {
	client client.Client
}

func NewWorkflowStarter(c client.Client) *WorkflowStarterImpl {
	return &WorkflowStarterImpl{client: c}
}

func (s *WorkflowStarterImpl) StartCreateOrderWorkflow(ctx context.Context, order appmodel.Order) (uuid.UUID, error) {
	options := client.StartWorkflowOptions{
		ID:        "order-" + uuid.New().String(),
		TaskQueue: TaskQueue,
	}
	input := CreateOrderWorkflowInput{Order: order}
	we, err := s.client.ExecuteWorkflow(ctx, options, CreateOrderWorkflow, input)
	if err != nil {
		return uuid.Nil, err
	}

	var result CreateOrderWorkflowResult
	err = we.Get(ctx, &result)
	if err != nil {
		return uuid.Nil, err
	}
	return result.OrderID, nil
}
