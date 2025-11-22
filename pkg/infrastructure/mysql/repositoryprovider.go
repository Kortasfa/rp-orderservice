package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"order/pkg/application/service"
	"order/pkg/domain/model"
	"order/pkg/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) OrderRepository(ctx context.Context) model.OrderRepository {
	return repository.NewOrderRepository(ctx, r.client)
}
