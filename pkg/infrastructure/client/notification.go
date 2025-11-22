package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	notificationapi "order/api/client/notificationinternal"
)

type NotificationClient struct {
	client notificationapi.NotificationInternalServiceClient
}

func NewNotificationClient(conn *grpc.ClientConn) *NotificationClient {
	return &NotificationClient{
		client: notificationapi.NewNotificationInternalServiceClient(conn),
	}
}

func (c *NotificationClient) SendNotification(ctx context.Context, userID uuid.UUID, message string) error {
	_, err := c.client.SendNotification(ctx, &notificationapi.SendNotificationRequest{
		UserID:  userID.String(),
		Message: message,
	})
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	return nil
}
