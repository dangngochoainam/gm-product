package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dangngochoainam/gm-core/apperror"
	productdtov1 "github.com/dangngochoainam/gm-core/application/product/v1/dto"
	"github.com/dangngochoainam/gm-product/internal/dto"
	"github.com/dangngochoainam/gopkg/commonhelper"
	"github.com/dangngochoainam/gopkg/copyhelper"
)

type fakeProductUsecase struct {
	listResp   *dto.ListProductsResponseDTO
	createResp *dto.CreateManyProductsResponseDTO
	err        error
}

func (f *fakeProductUsecase) ListProducts(context.Context, *dto.ListProductsRequestDTO) (*dto.ListProductsResponseDTO, error) {
	return f.listResp, f.err
}

func (f *fakeProductUsecase) CreateManyProducts(context.Context, *dto.CreateManyProductsRequestDTO) (*dto.CreateManyProductsResponseDTO, error) {
	return f.createResp, f.err
}

func newTestAPI(uc *fakeProductUsecase) *api {
	return NewAPI(copyhelper.NewPbConverter(), nil, uc).(*api)
}

func traceCtx() context.Context {
	return context.WithValue(context.Background(), commonhelper.TRACE_ID, "trace-1")
}

func TestListProductsReturnsErrorNotEnvelope(t *testing.T) {
	a := newTestAPI(&fakeProductUsecase{err: apperror.ErrBadRequest})

	resp, err := a.ListProducts(traceCtx(), &productdtov1.ListProductsRequest{Limit: 10})

	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}
	if !errors.Is(err, apperror.ErrBadRequest) {
		t.Errorf("err = %v, want wrapping ErrBadRequest", err)
	}
}

func TestCreateManyProductsReturnsErrorNotEnvelope(t *testing.T) {
	a := newTestAPI(&fakeProductUsecase{err: apperror.ErrBadRequest})

	resp, err := a.CreateManyProducts(traceCtx(), &productdtov1.CreateManyProductsRequest{
		Products: []*productdtov1.ProductItem{{Name: "n", Code: "c"}},
	})

	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}
	if !errors.Is(err, apperror.ErrBadRequest) {
		t.Errorf("err = %v, want wrapping ErrBadRequest", err)
	}
}

func TestListProductsSuccess(t *testing.T) {
	now := time.Date(2026, 10, 10, 8, 0, 0, 0, time.UTC)
	a := newTestAPI(&fakeProductUsecase{listResp: &dto.ListProductsResponseDTO{
		Products:   []*dto.ProductDetailDTO{{ID: "p1", Name: "n", Code: "c", CreatedAt: now, LastModifiedAt: now}},
		Pagination: &dto.BaseKeysetPaginationResponseDTO{HasNext: true, NextCursor: "c"},
	}})

	resp, err := a.ListProducts(traceCtx(), &productdtov1.ListProductsRequest{Limit: 10})

	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if len(resp.GetProducts()) != 1 || resp.GetProducts()[0].GetId() != "p1" {
		t.Errorf("products = %v, want one product p1", resp.GetProducts())
	}
	if resp.GetPagination().GetNextCursor() != "c" || !resp.GetPagination().GetHasNext() {
		t.Errorf("pagination = %v, want {HasNext: true, NextCursor: c}", resp.GetPagination())
	}
}
