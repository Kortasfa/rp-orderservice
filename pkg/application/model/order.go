package model

import "github.com/google/uuid"

type Order struct {
	UserID uuid.UUID
	Items  []OrderItem
}

type OrderItem struct {
	ProductID uuid.UUID
	Quantity  int
}
