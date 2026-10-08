package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/go-reading-01/internal/product"
)

// slowStore は、締め切りが来るまで返事をしない偽物の倉庫係。
type slowStore struct{}

func (slowStore) FindByID(ctx context.Context, id int64) (product.Product, error) {
	<-ctx.Done()
	return product.Product{}, ctx.Err()
}

func TestGetProduct_Timeout_Returns504(t *testing.T) {
	// 準備: 締め切り 0.05 秒の担当者に、返事をしない倉庫係を渡す
	svc := product.NewService(slowStore{}, 50*time.Millisecond)
	h := NewHandler(nil, svc)

	// 実行: サーバーを立てずに、GET /products/1 を1回だけ送ったことにする
	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	// 確認: 504 が返ったか
	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusGatewayTimeout, rec.Body.String())
	}
}
