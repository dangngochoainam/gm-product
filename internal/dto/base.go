package dto

const (
	SortOrderAsc  = "ASC"
	SortOrderDesc = "DESC"
)

type (
	BaseRequestDTO struct {
		TraceId string
	}
	BaseResponseDTO struct {
		IsProcessing bool
		IsError      bool
	}
	BasePaginationRequestDTO struct {
		Page  int32
		Limit int32
	}
	BasePaginationResponseDTO struct {
		Page  int32
		Limit int32
		Total int32
	}

	BaseCursorPaginationDTO struct {
		Cursor string
		Limit  int32
	}
	BaseCursorPaginationResponseDTO struct {
		HasNext bool
		Cursor  string
	}

	BaseKeysetPaginationRequestDTO struct {
		Limit     int32
		After     string // opaque cursor; empty means first page
		SortBy    string
		SortOrder string // "asc" or "desc"; defaults to "desc"
	}
	BaseKeysetPaginationResponseDTO struct {
		HasNext    bool
		NextCursor string
	}
)
