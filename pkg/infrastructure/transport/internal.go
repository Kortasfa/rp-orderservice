package transport

import (
	api "order/api/server/orderinternal"
)

func NewInternalAPI() api.OrderInternalServiceServer {
	return &internalAPI{}
}

type internalAPI struct {
	api.UnimplementedOrderInternalServiceServer
}
