package product

import (
	"context"
	"fmt"
	"time"

	"github.com/dangngochoainam/gm-product/internal/dto"
	"github.com/dangngochoainam/gopkg/copyhelper"
	"github.com/dangngochoainam/gopkg/gormhelper"
	"gorm.io/gorm"
)

// --- Input DTOs ---

type CreateProductInput struct {
	Name             string
	Code             string
	Price            float64
	Image            *string
	CreatedUser      *string
	LastModifiedUser *string
}

// --- Output DTOs ---

type ProductOutput struct {
	ID               string
	Name             string
	Code             string
	Price            float64
	Image            *string
	CreatedAt        time.Time
	CreatedUser      *string
	LastModifiedAt   time.Time
	LastModifiedUser *string
	DeletedAt        *time.Time
}

type CreateManyProductsOutput struct {
	Products []*ProductOutput
}

// ProductCursor is the keyset position used for stable pagination.
// It is serialized into the opaque cursor token returned to callers.
// SortBy records which field was active so the next page uses the same ordering.
type ProductCursor struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Name      string    `json:"name,omitempty"`
}

type ListProductsWithKeysetInput struct {
	Limit     int
	SortBy    string
	SortOrder string // "asc" or "desc"; defaults to "desc"
	// After is the decoded cursor from the previous page's last item.
	// A nil value means "start from the beginning".
	After *ProductCursor
}

func (i *ListProductsWithKeysetInput) Validate() error {
	if i.SortOrder != "" && i.SortOrder != dto.SortOrderAsc && i.SortOrder != dto.SortOrderDesc {
		return fmt.Errorf("invalid sort_order %q: must be %q, %q, or empty", i.SortOrder, dto.SortOrderAsc, dto.SortOrderDesc)
	}
	return nil
}

type ListProductsWithKeysetOutput struct {
	Products []*ProductOutput
	// HasNext is true when there is at least one more page after this one.
	HasNext bool
}

// --- Repository ---

type ProductRepository interface {
	gormhelper.Repository[ProductRepository]
	CreateMany(ctx context.Context, inputs []*CreateProductInput, batchSize int) (*CreateManyProductsOutput, error)
	ListWithKeyset(ctx context.Context, input *ListProductsWithKeysetInput) (*ListProductsWithKeysetOutput, error)
}

type productRepository struct {
	db        *gorm.DB
	converter copyhelper.ModelConverter
}

func NewProductRepository(gormDB *gorm.DB, converter copyhelper.ModelConverter) ProductRepository {
	return &productRepository{db: gormDB, converter: converter}
}

func (r *productRepository) WithTx(tx *gorm.DB) ProductRepository {
	return &productRepository{db: tx, converter: r.converter}
}
