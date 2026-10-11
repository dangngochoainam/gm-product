package usecase

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/dangngochoainam/gm-core/apperror"
	"github.com/dangngochoainam/gm-product/internal/dto"
	productrepo "github.com/dangngochoainam/gm-product/internal/repository/product"
	"github.com/dangngochoainam/gopkg/copyhelper"
	"github.com/dangngochoainam/gopkg/loghelper"
)

func TestMain(m *testing.M) {
	// The use case logs repository failures through the package-level logger.
	if err := loghelper.InitZapLogger("prod", "usecase-test"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// fakeProductRepository implements only the methods the use case calls; the
// embedded interface is nil, so any other call panics.
type fakeProductRepository struct {
	productrepo.ProductRepository
	listOut  *productrepo.ListProductsWithKeysetOutput
	listErr  error
	gotInput *productrepo.ListProductsWithKeysetInput
}

func (f *fakeProductRepository) ListWithKeyset(_ context.Context, input *productrepo.ListProductsWithKeysetInput) (*productrepo.ListProductsWithKeysetOutput, error) {
	f.gotInput = input
	return f.listOut, f.listErr
}

func listProducts(t *testing.T, repo *fakeProductRepository, req dto.BaseKeysetPaginationRequestDTO) error {
	t.Helper()
	uc := NewProductUseCase(repo, copyhelper.NewModelConverter())
	_, err := uc.ListProducts(context.Background(), &dto.ListProductsRequestDTO{BaseKeysetPaginationRequestDTO: req})
	return err
}

func assertBadRequest(t *testing.T, err error, wantMessage string) {
	t.Helper()
	if !errors.Is(err, apperror.ErrBadRequest) {
		t.Fatalf("err = %v, want wrapping ErrBadRequest", err)
	}
	if err.Error() != wantMessage {
		t.Errorf("err.Error() = %q, want %q", err.Error(), wantMessage)
	}
}

func TestListProductsRejectsSortBy(t *testing.T) {
	err := listProducts(t, &fakeProductRepository{}, dto.BaseKeysetPaginationRequestDTO{Limit: 10, SortBy: "price"})
	assertBadRequest(t, err, "bad request: sortBy must be created_at or name")
}

func TestListProductsRejectsSortOrder(t *testing.T) {
	err := listProducts(t, &fakeProductRepository{}, dto.BaseKeysetPaginationRequestDTO{Limit: 10, SortOrder: "up"})
	assertBadRequest(t, err, "bad request: sortOrder must be asc or desc")
}

func TestListProductsRejectsCursor(t *testing.T) {
	err := listProducts(t, &fakeProductRepository{}, dto.BaseKeysetPaginationRequestDTO{Limit: 10, After: "!!!"})
	assertBadRequest(t, err, "bad request: after is not a valid cursor")
}

func TestListProductsRepositoryError(t *testing.T) {
	err := listProducts(t, &fakeProductRepository{listErr: errors.New("db down")}, dto.BaseKeysetPaginationRequestDTO{Limit: 10})
	if !errors.Is(err, apperror.ErrInternalServerError) {
		t.Fatalf("err = %v, want wrapping ErrInternalServerError", err)
	}
}

// assertSortPassed checks the sort the use case handed to the repository.
func assertSortPassed(t *testing.T, err error, repo *fakeProductRepository, wantSortBy, wantSortOrder string) {
	t.Helper()
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if repo.gotInput == nil {
		t.Fatal("repository was not called")
	}
	if repo.gotInput.SortBy != wantSortBy || repo.gotInput.SortOrder != wantSortOrder {
		t.Errorf("repository got sort {%q, %q}, want {%q, %q}", repo.gotInput.SortBy, repo.gotInput.SortOrder, wantSortBy, wantSortOrder)
	}
}

func TestListProductsAcceptsValidSort(t *testing.T) {
	repo := &fakeProductRepository{listOut: &productrepo.ListProductsWithKeysetOutput{}}

	err := listProducts(t, repo, dto.BaseKeysetPaginationRequestDTO{Limit: 10, SortBy: "name", SortOrder: "asc"})

	assertSortPassed(t, err, repo, "name", "ASC")
}

func TestListProductsDefaultsSortOrder(t *testing.T) {
	repo := &fakeProductRepository{listOut: &productrepo.ListProductsWithKeysetOutput{}}

	err := listProducts(t, repo, dto.BaseKeysetPaginationRequestDTO{Limit: 10})

	assertSortPassed(t, err, repo, "", "DESC")
}
