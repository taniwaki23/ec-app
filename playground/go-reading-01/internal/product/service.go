package product

import (
	"context"
	"time"
)

// Store は Service が必要とする「商品を取ってくる」能力だけを表す。
// 使う側(Service)がこのインターフェースを宣言し、Repository は何も書かずにこれを満たす。
type Store interface {
	FindByID(ctx context.Context, id int64) (Product, error)
}

type Service struct {
	store   Store
	timeout time.Duration
}

func NewService(store Store, timeout time.Duration) *Service {
	return &Service{store: store, timeout: timeout}
}

func (s *Service) Get(ctx context.Context, id int64) (Product, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.store.FindByID(ctx, id)
}
