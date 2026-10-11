package api

import (
	"context"

	productdtov1 "github.com/dangngochoainam/gm-core/application/product/v1/dto"
	"github.com/dangngochoainam/gm-product/internal/dto"
	"github.com/dangngochoainam/gopkg/commonhelper"
)

func (a *api) ListProducts(ctx context.Context, request *productdtov1.ListProductsRequest) (*productdtov1.ListProductsResponse, error) {
	reqDTO := &dto.ListProductsRequestDTO{BaseRequestDTO: dto.BaseRequestDTO{TraceId: ctx.Value(commonhelper.TRACE_ID).(string)}}

	a.pbConverter.FromPb(reqDTO, request)

	respDTO, err := a.productUsecase.ListProducts(ctx, reqDTO)
	if err != nil {
		return nil, err
	}

	resp := &productdtov1.ListProductsResponse{}
	a.pbConverter.ToPb(resp, respDTO)

	return resp, nil
}

func (a *api) CreateManyProducts(ctx context.Context, request *productdtov1.CreateManyProductsRequest) (*productdtov1.CreateManyProductsResponse, error) {
	reqDTO := &dto.CreateManyProductsRequestDTO{BaseRequestDTO: dto.BaseRequestDTO{TraceId: ctx.Value(commonhelper.TRACE_ID).(string)}}

	a.pbConverter.FromPb(reqDTO, request)

	respDTO, err := a.productUsecase.CreateManyProducts(ctx, reqDTO)
	if err != nil {
		return nil, err
	}

	resp := &productdtov1.CreateManyProductsResponse{}
	a.pbConverter.ToPb(resp, respDTO)
	return resp, nil
}
