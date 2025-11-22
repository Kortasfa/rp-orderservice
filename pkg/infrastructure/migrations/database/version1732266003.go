package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion1732266003(client mysql.ClientContext) migrator.Migration {
	return &version1732266003{
		client: client,
	}
}

type version1732266003 struct {
	client mysql.ClientContext
}

func (v version1732266003) Version() int64 {
	return 1732266003
}

func (v version1732266003) Description() string {
	return "Create 'orders' and 'order_items' tables"
}

func (v version1732266003) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
CREATE TABLE orders
(
    order_id    VARCHAR(64)  NOT NULL,
    user_id     VARCHAR(64)  NOT NULL,
    status      VARCHAR(32)  NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    created_at  DATETIME     NOT NULL,
    updated_at  DATETIME     NOT NULL,
    PRIMARY KEY (order_id),
    INDEX idx_user_id (user_id)
)
    ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
`)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = v.client.ExecContext(ctx, `
CREATE TABLE order_items
(
    item_id     VARCHAR(64)  NOT NULL,
    order_id    VARCHAR(64)  NOT NULL,
    product_id  VARCHAR(64)  NOT NULL,
    quantity    INT          NOT NULL,
    price       DECIMAL(10, 2) NOT NULL,
    PRIMARY KEY (item_id),
    INDEX idx_order_id (order_id)
)
    ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
`)
	return errors.WithStack(err)
}
