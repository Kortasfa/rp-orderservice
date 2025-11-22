package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	productapi "order/api/client/productinternal"
)

type ProductClient struct {
	client productapi.ProductInternalServiceClient
}

func NewProductClient(conn *grpc.ClientConn) *ProductClient {
	return &ProductClient{
		client: productapi.NewProductInternalServiceClient(conn),
	}
}

func (c *ProductClient) GetPrice(ctx context.Context, productID uuid.UUID) (float64, error) {
	resp, err := c.client.FindProduct(ctx, &productapi.FindProductRequest{
		ProductID: productID.String(),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to find product: %w", err)
	}
	if resp.Product == nil {
		return 0, fmt.Errorf("product not found")
	}
	return resp.Product.Price, nil
}
