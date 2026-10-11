//go:build integration

package api

import (
	"context"
	"net"
	"os"
	"testing"

	productdtov1 "github.com/dangngochoainam/gm-core/application/product/v1/dto"
	productsvcv1 "github.com/dangngochoainam/gm-core/application/product/v1/service"
	"github.com/dangngochoainam/gm-product/configs"
	"github.com/dangngochoainam/gm-product/internal/repository/product"
	"github.com/dangngochoainam/gm-product/internal/usecase"
	"github.com/dangngochoainam/gopkg/copyhelper"
	"github.com/dangngochoainam/gopkg/gormhelper"
	"github.com/dangngochoainam/gopkg/loghelper"
	"github.com/dangngochoainam/gopkg/serverhelper"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gorm.io/gorm"
)

const integrationDomain = "gm-product"

var postgresEnv = []string{
	"DATABASE_POSTGRES__HOST",
	"DATABASE_POSTGRES__PORT",
	"DATABASE_POSTGRES__DATABASE",
	"DATABASE_POSTGRES__SCHEMA",
	"DATABASE_POSTGRES__USERNAME",
	"DATABASE_POSTGRES__PASSWORD",
}

func TestMain(m *testing.M) {
	// The server interceptors and the use case log through the package-level logger.
	if err := loghelper.InitZapLogger("prod", "api-integration-test"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// openDB connects to the Postgres named by the DATABASE_POSTGRES__* variables,
// the same ones the service reads.
func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	for _, name := range postgresEnv {
		if os.Getenv(name) == "" {
			t.Fatalf("%s is not set", name)
		}
	}
	cfg, err := configs.LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	pg := cfg.DatabasePostgres
	db, err := gormhelper.NewGormHelper(&gormhelper.GormHelperConfig{
		Host:     pg.Host,
		Port:     pg.Port,
		Database: pg.Database,
		Schema:   pg.Schema,
		Username: pg.Username,
		Password: pg.Password,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	return db
}

// startServer serves the real API, use case and repository on db behind the
// interceptor chain gm-product runs in production, and returns a client for it.
func startServer(t *testing.T, db *gorm.DB) productsvcv1.ProductServiceClient {
	t.Helper()
	repo := product.NewProductRepository(db, copyhelper.NewModelConverter())
	uc := usecase.NewProductUseCase(repo, copyhelper.NewModelConverter())

	server := serverhelper.CreateGrpcServer(serverhelper.WithServiceName(integrationDomain))
	productsvcv1.RegisterProductServiceServer(server, NewAPI(copyhelper.NewPbConverter(), nil, uc))

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() { _ = server.Serve(lis) }()
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return productsvcv1.NewProductServiceClient(conn)
}

// assertGRPCError checks the status code and the ErrorInfo every gm-product
// error must carry, and returns the status for further checks.
func assertGRPCError(t *testing.T, err error, wantCode codes.Code, wantReason string) *status.Status {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %v", wantCode)
	}
	st := status.Convert(err)
	if st.Code() != wantCode {
		t.Errorf("code = %v, want %v (message %q)", st.Code(), wantCode, st.Message())
	}
	var info *errdetails.ErrorInfo
	for _, d := range st.Details() {
		if i, ok := d.(*errdetails.ErrorInfo); ok {
			info = i
		}
	}
	if info == nil {
		t.Fatalf("no ErrorInfo in details: %v", st.Details())
	}
	if info.GetReason() != wantReason || info.GetDomain() != integrationDomain {
		t.Errorf("ErrorInfo = {%q, %q}, want {%q, %q}", info.GetReason(), info.GetDomain(), wantReason, integrationDomain)
	}
	return st
}

func fieldViolations(st *status.Status) []*errdetails.BadRequest_FieldViolation {
	for _, d := range st.Details() {
		if br, ok := d.(*errdetails.BadRequest); ok {
			return br.GetFieldViolations()
		}
	}
	return nil
}

func assertLimitRejected(t *testing.T, client productsvcv1.ProductServiceClient, limit int32) {
	t.Helper()
	_, err := client.ListProducts(context.Background(), &productdtov1.ListProductsRequest{Limit: limit})

	st := assertGRPCError(t, err, codes.InvalidArgument, "VALIDATION_FAILED")
	fv := fieldViolations(st)
	if len(fv) == 0 || fv[0].GetField() != "limit" {
		t.Errorf("field violations = %v, want first field limit", fv)
	}
}

func TestIntegrationListProductsSuccess(t *testing.T) {
	client := startServer(t, openDB(t))

	resp, err := client.ListProducts(context.Background(), &productdtov1.ListProductsRequest{Limit: 2})

	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if resp.GetPagination() == nil {
		t.Errorf("pagination missing: %v", resp)
	}
	if n := len(resp.GetProducts()); n > 2 {
		t.Errorf("products = %d, want at most 2", n)
	}
	fields := resp.ProtoReflect().Descriptor().Fields()
	for _, name := range []protoreflect.Name{"trace_id", "status_code", "reason_code", "reason_message", "time_ms"} {
		if fields.ByName(name) != nil {
			t.Errorf("envelope field %q must be gone from ListProductsResponse", name)
		}
	}
}

func TestIntegrationCreateManyProductsEmpty(t *testing.T) {
	client := startServer(t, openDB(t))

	_, err := client.CreateManyProducts(context.Background(), &productdtov1.CreateManyProductsRequest{})

	st := assertGRPCError(t, err, codes.InvalidArgument, "VALIDATION_FAILED")
	fv := fieldViolations(st)
	if len(fv) == 0 || fv[0].GetField() != "products" {
		t.Errorf("field violations = %v, want first field products", fv)
	}
}

func TestIntegrationListProductsLimitBounds(t *testing.T) {
	client := startServer(t, openDB(t))

	for _, limit := range []int32{1, 100} {
		if _, err := client.ListProducts(context.Background(), &productdtov1.ListProductsRequest{Limit: limit}); err != nil {
			t.Errorf("limit %d: err = %v, want nil", limit, err)
		}
	}
}

func TestIntegrationListProductsLimitBelowMin(t *testing.T) {
	assertLimitRejected(t, startServer(t, openDB(t)), 0)
}

func TestIntegrationListProductsLimitAboveMax(t *testing.T) {
	assertLimitRejected(t, startServer(t, openDB(t)), 101)
}

func TestIntegrationListProductsInvalidSortBy(t *testing.T) {
	client := startServer(t, openDB(t))

	_, err := client.ListProducts(context.Background(), &productdtov1.ListProductsRequest{Limit: 2, SortBy: "price"})

	st := assertGRPCError(t, err, codes.InvalidArgument, "BAD_REQUEST")
	if want := "bad request: sortBy must be created_at or name"; st.Message() != want {
		t.Errorf("message = %q, want %q", st.Message(), want)
	}
}

func TestIntegrationListProductsDBDown(t *testing.T) {
	db := openDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql.DB: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	client := startServer(t, db)

	_, err = client.ListProducts(context.Background(), &productdtov1.ListProductsRequest{Limit: 2})

	assertGRPCError(t, err, codes.Internal, "INTERNAL_SERVER_ERROR")
}
