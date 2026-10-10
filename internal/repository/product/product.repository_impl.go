package product

import (
	"context"

	"github.com/dangngochoainam/gm-product/db/entities"
	"github.com/dangngochoainam/gm-product/internal/dto"
)

func (r *productRepository) CreateMany(ctx context.Context, inputs []*CreateProductInput, batchSize int) (*CreateManyProductsOutput, error) {
	records := make([]*entities.Product, len(inputs))
	for i, input := range inputs {
		e := &entities.Product{}
		r.converter.ToModel(e, input)
		records[i] = e
	}

	if err := r.db.WithContext(ctx).CreateInBatches(records, batchSize).Error; err != nil {
		return nil, err
	}

	outputs := make([]*ProductOutput, len(records))
	for i, e := range records {
		outputs[i] = r.toProductOutput(e)
	}

	return &CreateManyProductsOutput{Products: outputs}, nil
}

func (r *productRepository) ListWithKeyset(ctx context.Context, input *ListProductsWithKeysetInput) (*ListProductsWithKeysetOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Fetch one extra row to decide whether a next page exists without COUNT(*).
	fetchLimit := input.Limit + 1

	db := r.db.WithContext(ctx).Model(&entities.Product{}).Limit(fetchLimit)

	// cursor comparison operator: ASC order pages forward with >, DESC with <
	cursorOp := "<"
	if input.SortOrder == dto.SortOrderAsc {
		cursorOp = ">"
	}

	switch input.SortBy {
	case "name":
		db = db.Order("name " + input.SortOrder + ", id " + input.SortOrder)
		if input.After != nil {
			db = db.Where("(name, id) "+cursorOp+" (?, ?)", input.After.Name, input.After.ID)
		}
	default: // "created_at" or ""
		db = db.Order("created_at " + input.SortOrder + ", id " + input.SortOrder)
		if input.After != nil {
			db = db.Where("(created_at, id) "+cursorOp+" (?, ?)", input.After.CreatedAt, input.After.ID)
		}
	}

	var items []entities.Product
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}

	hasNext := len(items) > input.Limit
	if hasNext {
		items = items[:input.Limit]
	}

	outputs := make([]*ProductOutput, len(items))
	for i := range items {
		outputs[i] = r.toProductOutput(&items[i])
	}

	return &ListProductsWithKeysetOutput{
		Products: outputs,
		HasNext:  hasNext,
	}, nil
}

func (r *productRepository) toProductOutput(e *entities.Product) *ProductOutput {
	output := &ProductOutput{}
	r.converter.FromModel(output, e)
	if e.DeletedAt.Valid {
		output.DeletedAt = &e.DeletedAt.Time
	}
	return output
}
