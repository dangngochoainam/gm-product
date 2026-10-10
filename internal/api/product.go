package api

import (
	"context"

	productdtov1 "github.com/dangngochoainam/gm-core/application/product/v1/dto"
	"github.com/dangngochoainam/gm-product/internal/dto"
	"github.com/dangngochoainam/gopkg/commonhelper"
	"github.com/dangngochoainam/gopkg/responsehelper"
)

func (a *api) ListProducts(ctx context.Context, request *productdtov1.ListProductsRequest) (*productdtov1.ListProductsResponse, error) {
	var (
		reqDTO  = &dto.ListProductsRequestDTO{BaseRequestDTO: dto.BaseRequestDTO{TraceId: ctx.Value(commonhelper.TRACE_ID).(string)}}
		respDTO *dto.ListProductsResponseDTO
	)

	a.pbConverter.FromPb(reqDTO, request)

	resp := &productdtov1.ListProductsResponse{
		TraceId:    reqDTO.TraceId,
		StatusCode: responsehelper.StatusCodeAccept,
	}

	respDTO, err := a.productUsecase.ListProducts(ctx, reqDTO)
	if err != nil {
		errResp := &productdtov1.ListProductsResponse{
			TraceId:       reqDTO.TraceId,
			StatusCode:    responsehelper.StatusCodeReject,
			ReasonCode:    responsehelper.ParseError(err).Code(),
			ReasonMessage: responsehelper.ParseError(err).Message(),
		}
		if respDTO != nil && respDTO.IsError {
			errResp.StatusCode = responsehelper.StatusCodeError
		}
		return errResp, nil
	}

	if respDTO != nil && respDTO.IsProcessing {
		resp.StatusCode = responsehelper.StatusCodeProcessing
	}

	a.pbConverter.ToPb(resp, respDTO)

	return resp, nil
}

func (a *api) CreateManyProducts(ctx context.Context, request *productdtov1.CreateManyProductsRequest) (*productdtov1.CreateManyProductsResponse, error) {
	var (
		reqDTO  = &dto.CreateManyProductsRequestDTO{BaseRequestDTO: dto.BaseRequestDTO{TraceId: ctx.Value(commonhelper.TRACE_ID).(string)}}
		respDTO *dto.CreateManyProductsResponseDTO
	)

	a.pbConverter.FromPb(reqDTO, request)

	resp := &productdtov1.CreateManyProductsResponse{
		TraceId:    reqDTO.TraceId,
		StatusCode: responsehelper.StatusCodeAccept,
	}

	respDTO, err := a.productUsecase.CreateManyProducts(ctx, reqDTO)
	if err != nil {
		resp := &productdtov1.CreateManyProductsResponse{
			TraceId:       reqDTO.TraceId,
			StatusCode:    responsehelper.StatusCodeReject,
			ReasonCode:    responsehelper.ParseError(err).Code(),
			ReasonMessage: responsehelper.ParseError(err).Message(),
		}
		if respDTO != nil && respDTO.IsError {
			resp.StatusCode = responsehelper.StatusCodeError
		}
		return resp, nil
	}

	if respDTO != nil && respDTO.IsProcessing {
		resp.StatusCode = responsehelper.StatusCodeProcessing
	}

	a.pbConverter.ToPb(resp, respDTO)
	return resp, nil
}
