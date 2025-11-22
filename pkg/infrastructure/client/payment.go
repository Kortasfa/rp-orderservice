package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	paymentapi "order/api/client/paymentserviceinternal"
)

type PaymentClient struct {
	client paymentapi.PaymentServiceInternalClient
}

func NewPaymentClient(conn *grpc.ClientConn) *PaymentClient {
	return &PaymentClient{
		client: paymentapi.NewPaymentServiceInternalClient(conn),
	}
}

func (c *PaymentClient) ProcessPayment(ctx context.Context, userID, orderID uuid.UUID, amount float64) error {
	_, err := c.client.ProcessPayment(ctx, &paymentapi.ProcessPaymentRequest{
		UserID:  userID.String(),
		OrderID: orderID.String(),
		Amount:  amount,
	})
	if err != nil {
		return fmt.Errorf("failed to process payment: %w", err)
	}
	return nil
}
