package queue

import (
	"context"
	"github.com/arandich/marketplace-api/internal/model"
)

type IOrderQueue interface {
	PublishOrder(ctx context.Context, order model.OrderMsg) error
}
