package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/dangngochoainam/gm-core/apperror"
	"github.com/dangngochoainam/gm-product/internal/dto"
	productrepo "github.com/dangngochoainam/gm-product/internal/repository/product"
	"github.com/dangngochoainam/gopkg/copyhelper"
	"github.com/dangngochoainam/gopkg/cursorhelper"
	"github.com/dangngochoainam/gopkg/loghelper"
)

type product struct {
	productRepository productrepo.ProductRepository
	modelConverter    copyhelper.ModelConverter
}

func NewProductUseCase(productRepository productrepo.ProductRepository, modelConverter copyhelper.ModelConverter) Product {
	return &product{productRepository: productRepository, modelConverter: modelConverter}
}

type Product interface {
	CreateManyProducts(ctx context.Context, input *dto.CreateManyProductsRequestDTO) (*dto.CreateManyProductsResponseDTO, error)
	ListProducts(ctx context.Context, input *dto.ListProductsRequestDTO) (*dto.ListProductsResponseDTO, error)
}

func (u *product) CreateManyProducts(ctx context.Context, input *dto.CreateManyProductsRequestDTO) (*dto.CreateManyProductsResponseDTO, error) {
	result := &dto.CreateManyProductsResponseDTO{}

	inputs := make([]*productrepo.CreateProductInput, len(input.Products))
	for i, p := range input.Products {
		image := p.Image
		inputs[i] = &productrepo.CreateProductInput{
			Name:  p.Name,
			Code:  p.Code,
			Price: p.Price,
			Image: &image,
		}
	}

	created, err := u.productRepository.CreateMany(ctx, inputs, 10)
	if err != nil {
		loghelper.Logger.WithContext(ctx).Errorf("Failed while creating products: %v", err)
		return nil, apperror.ErrInternalServerError
	}

	result.Products = make([]*dto.ProductDetailDTO, len(created.Products))
	for i, p := range created.Products {
		detail := &dto.ProductDetailDTO{}
		u.modelConverter.FromModel(detail, p)
		if p.Image != nil {
			detail.Image = *p.Image
		}
		result.Products[i] = detail
	}

	return result, nil
}

func (u *product) ListProducts(ctx context.Context, input *dto.ListProductsRequestDTO) (*dto.ListProductsResponseDTO, error) {
	result := &dto.ListProductsResponseDTO{
		Pagination: &dto.BaseKeysetPaginationResponseDTO{},
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	sortBy := input.SortBy
	// TODO: I think sortBy will based on each api, each api will allow sortBy different
	if sortBy != "" && sortBy != "created_at" && sortBy != "name" {
		return nil, fmt.Errorf("%w: sortBy must be created_at or name", apperror.ErrBadRequest)
	}

	sortOrder := strings.ToUpper(input.SortOrder)
	if sortOrder != "" && sortOrder != dto.SortOrderAsc && sortOrder != dto.SortOrderDesc {
		return nil, fmt.Errorf("%w: sortOrder must be asc or desc", apperror.ErrBadRequest)
	}
	if sortOrder == "" {
		sortOrder = dto.SortOrderDesc
	}

	var after *productrepo.ProductCursor
	if input.After != "" {
		cursor, err := cursorhelper.Decode[productrepo.ProductCursor](input.After)
		if err != nil {
			return nil, fmt.Errorf("%w: after is not a valid cursor", apperror.ErrBadRequest)
		}
		after = cursor
	}

	out, err := u.productRepository.ListWithKeyset(ctx, &productrepo.ListProductsWithKeysetInput{
		Limit:     int(limit),
		SortBy:    sortBy,
		SortOrder: sortOrder,
		After:     after,
	})
	if err != nil {
		loghelper.Logger.WithContext(ctx).Errorf("Failed while listing products: %v", err)
		return nil, apperror.ErrInternalServerError
	}

	result.Products = make([]*dto.ProductDetailDTO, len(out.Products))
	for i, p := range out.Products {
		detail := &dto.ProductDetailDTO{}
		u.modelConverter.FromModel(detail, p)
		if p.Image != nil {
			detail.Image = *p.Image
		}
		result.Products[i] = detail
	}

	u.modelConverter.FromModel(result.Pagination, out)

	// NextCursor is derived (encoded from the last item), not a direct field copy.
	if out.HasNext && len(out.Products) > 0 {
		last := out.Products[len(out.Products)-1]
		nextCursor, err := cursorhelper.Encode(productrepo.ProductCursor{
			ID:        last.ID,
			CreatedAt: last.CreatedAt,
			Name:      last.Name,
		})
		if err != nil {
			loghelper.Logger.WithContext(ctx).Errorf("Failed while encoding cursor: %v", err)
			return nil, apperror.ErrInternalServerError
		}
		result.Pagination.NextCursor = nextCursor
	}

	return result, nil
}
