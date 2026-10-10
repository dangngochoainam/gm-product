package dto

import "time"

type CreateManyProductsRequestDTO struct {
	BaseRequestDTO
	Products []*CreateProductItemDTO
}

type CreateProductItemDTO struct {
	Name  string
	Code  string
	Price float64
	Image string
}

type CreateManyProductsResponseDTO struct {
	BaseResponseDTO
	Products []*ProductDetailDTO
}

type ProductDetailDTO struct {
	ID             string
	Name           string
	Code           string
	Price          float64
	Image          string
	CreatedAt      time.Time
	LastModifiedAt time.Time
}

type ListProductsRequestDTO struct {
	BaseRequestDTO
	BaseKeysetPaginationRequestDTO
}

type ListProductsResponseDTO struct {
	BaseResponseDTO
	Products   []*ProductDetailDTO
	Pagination *BaseKeysetPaginationResponseDTO
}
