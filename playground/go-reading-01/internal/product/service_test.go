package product

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeStore は DB を使わない Store。FindByID を持つので Store を満たす。
type fakeStore struct {
	product Product
	err     error
	delay   time.Duration
}

func (f *fakeStore) FindByID(ctx context.Context, id int64) (Product, error) {
	select {
	case <-time.After(f.delay):
		return f.product, f.err
	case <-ctx.Done():
		return Product{}, ctx.Err()
	}
}

func TestGet_ReturnsProduct(t *testing.T) {
	want := Product{ID: 1, Name: "テスト商品", Price: 1000, Stock: 3}
	s := NewService(&fakeStore{product: want}, time.Second)

	got, err := s.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGet_TimesOut(t *testing.T) {
	s := NewService(&fakeStore{delay: 2 * time.Second}, 100*time.Millisecond)

	start := time.Now()
	_, err := s.Get(context.Background(), 1)
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want context.DeadlineExceeded", err)
	}
	t.Logf("gave up after %v (store would have taken 2s)", elapsed)
}
