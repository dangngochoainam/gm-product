# CLAUDE.md

File này hướng dẫn Claude Code (claude.ai/code) khi làm việc với code trong repo này.

## Stack

- Module: `github.com/dangngochoainam/gm-product` · Go 1.24
- gRPC v1.76 + grpc-gateway v2.27 · GORM v1.31 (PostgreSQL) · Uber Dig v1.19 · Zap v1.27
- RabbitMQ (amqp091 v1.11) · protovalidate v1.1

## Lệnh

```bash
# Môi trường dev bằng Docker
make start              # docker compose -f docker-compose.debug.yaml up -d
make attach             # Vào shell của container debug đang chạy
make mod                # go mod tidy && go mod vendor

# Chạy
make run                # go run cmd/server/main.go
make run-watch          # Live reload qua CompileDaemon (Delve ở port 2302)

# Sinh code (proto nằm ở gm-core, không nằm ở đây — repo này không có target gen-proto)
cd ../gm-core && make gen-proto   # chạy trên host: cần buf, image này không mang theo
make mod                # quay lại đây, vendor lại để code sinh ra mới được nhận

# Kiểm tra (config lint ở .golangci.yml)
make fmt                # sửa format: gofmt + goimports
make lint               # golangci-lint v2.14.0 — lint toàn repo
make build              # go build ./...
make verify             # lint + build, không chạy test
make test               # go test ./... — unit test, không tag
```

## Kiến trúc

Clean architecture 4 layer, nối với nhau bằng dependency injection của Uber Dig (`internal/diregistry/`):

```
gRPC → API (internal/api/) → Use Case (internal/usecase/) → Repository (internal/repository/) → DB (GORM + PostgreSQL)
```

**Chỉ gRPC.** Service này không phục vụ HTTP public. REST từ bên ngoài kết thúc ở gm-gateway, nơi
dịch nó sang gRPC dựa trên annotation `google.api.http` trong gm-core. Ở đây không có grpc-gateway
mux, không CORS, không self-dial qua loopback, và thêm lại một cái như vậy là mở cho caller một cửa
thứ hai đi vòng qua mọi middleware của gateway. Lý do được ghi lại trong
[`docs/service-ports.md`](docs/service-ports.md).

Listener HTTP duy nhất là **admin port**: `/healthz`, `/readyz` và `/metrics`, dành cho kubelet và
Prometheus. Không bao giờ trỏ ingress vào nó.

**Chuỗi middleware** (áp cho mọi request, theo thứ tự):
1. Inject trace ID
2. Validate protobuf (`buf.build/protovalidate`)
3. Log payload request/response (zap)
4. Recover panic

### Trách nhiệm từng layer

| Layer | Vị trí | Vai trò |
|-------|--------|---------|
| Entry point | `cmd/server/main.go` | Dựng DI container, khởi động gRPC server và admin listener, drain khi nhận SIGTERM |
| API | `internal/api/` | Implement gRPC handler; convert protobuf ↔ DTO |
| Use Case | `internal/usecase/` | Logic nghiệp vụ; điều phối các repository |
| Repository | `internal/repository/` | Truy cập dữ liệu qua GORM |
| Entities | `db/entities/` | Định nghĩa model GORM; là nguồn cho migration schema |

### Package quan trọng

Helper dùng chung nằm ở repo ngoài [`gopkg`](https://github.com/dangngochoainam/gopkg) (mount ở local qua `replace` trong `go.mod`). Xem repo đó để có tài liệu package mới nhất.

### Quy trình protobuf → Go

Repo này **không có `proto/` và không có `pb/`**. Contract của product nằm trong `gm-core` tại
`../gm-core/proto/product/v1/`, chia thành `dto/product.dto.proto` (package `product.v1.dto`) và
`service/product.service.proto` (package `product.v1.service`). Import nửa nào mà file đó cần:

```go
productdtov1 "github.com/dangngochoainam/gm-core/application/product/v1/dto"      // message
productsvcv1 "github.com/dangngochoainam/gm-core/application/product/v1/service"  // server, client, handler
```

`internal/api/product.go` chỉ dựng message nên chỉ import `dto`. `api.go` và `cmd/server/main.go`
cần cả hai. Service gRPC tên là `ProductService`, nên các symbol sinh ra là `ProductServiceServer`,
`RegisterProductServiceServer` và `RegisterProductServiceHandler`.

1. Sửa `../gm-core/proto/product/v1/dto/product.dto.proto` (message) hoặc
   `service/product.service.proto` (rpc)
2. `cd ../gm-core && make gen-proto` → buf regenerate `../gm-core/application/product/v1/`.
   Repo này không có target `gen-proto`: không bao giờ generate gì vào đây. Code sinh ra không
   được commit, nên checkout gm-core mới cũng phải chạy bước này trước `make mod`
3. `make mod` ở đây → tidy và vendor lại để thấy được thay đổi
4. Implement hoặc cập nhật handler trong `internal/api/`

`gm-core` được dùng theo path, giống `gopkg`: `replace github.com/dangngochoainam/gm-core => ../gm-core`.
Container debug bind-mount nó tại `/gm-core` để replace đó resolve được từ `/app`.
Xem `../gm-core/CLAUDE.md` để biết quy ước đặt tên file và package.

### Quy trình database (Atlas + GORM)

Schema suy ra từ struct entity của GORM (`db/entities/`): **Product**, **Job** (cả hai embed
**BaseModel** với các field audit và soft delete).

1. Thêm/sửa struct entity GORM trong `db/entities/`
2. Đăng ký entity mới trong `db/golang-migrate/loader/main.go`
3. Sinh migration: `cd db/golang-migrate && ./shmake diff <migration_name>`
4. Áp migration: `./shmake apply`

### Dependency injection

Mọi dependency đăng ký trong `internal/diregistry/diregistry.go` bằng Uber Dig, theo thứ tự: Config → ModelConverter + PbConverter + HttpClientHelper → GormHelper → MinioStorage → ProductRepository + JobRepository → ProductUseCase + ExportUseCase + JobUseCase → API. Khi thêm service/repository/use case mới, đăng ký constructor của nó ở đó và thêm nó làm tham số cho constructor tiêu thụ.

**Pattern: luôn inject qua tham số constructor — không bao giờ khởi tạo trực tiếp bên trong constructor.**

```go
// SAI — hard-code, đi vòng qua DI
func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepository{converter: copyhelper.NewModelConverter()}
}

// ĐÚNG — nhận qua tham số để Dig inject vào
func NewProductRepository(db *gorm.DB, converter copyhelper.ModelConverter) ProductRepository {
    return &productRepository{db: db, converter: converter}
}
```

### Copy giữa các struct

Dùng helper copier của project khi copy struct vào **output/result** trả về từ một hàm. Quy tắc này
**không** áp dụng khi dựng struct input truyền xuống layer kế tiếp — ở đó gán field thủ công mới là đúng.

| Helper | Method | Dùng khi nào |
|--------|--------|--------------|
| `PbConverter` | `FromPb(dst, src)` | Protobuf request → DTO — ở **đầu** handler API |
| `PbConverter` | `ToPb(dst, src)` | DTO → Protobuf response — ở **cuối** handler API |
| `ModelConverter` | `FromModel(dst, src)` | Entity DB/output repo → DTO — ở **cuối** hàm use case / repository |
| `ModelConverter` | `ToModel(dst, src)` | DTO → entity DB — ở **cuối** hàm repository |

```go
// SAI — gán field thủ công khi copy vào giá trị trả về
reqDTO.Limit = request.GetLimit()
reqDTO.After = request.GetAfter()

// ĐÚNG — dùng converter ở đầu handler API
a.pbConverter.FromPb(reqDTO, request)
```

Field không map được theo tên (ví dụ phải unwrap con trỏ, lệch kiểu) là lý do hợp lệ duy nhất để
gán thủ công trên một bản copy vào giá trị trả về, và chỉ cho đúng những field đó, sau khi đã gọi converter.

## Cấu hình

Copy `.env.example` thành `.env`. Viper dùng `__` làm dấu phân cách key, nên field config lồng nhau
map sang biến môi trường như sau:

| Field config | Biến môi trường |
|---|---|
| `database_postgres.host` | `DATABASE_POSTGRES__HOST` |
| `grpc_port` / `admin_port` | `GRPC_PORT` / `ADMIN_PORT` |
| `shutdown_drain_delay` / `shutdown_timeout` | `SHUTDOWN_DRAIN_DELAY` / `SHUTDOWN_TIMEOUT` |

Config mặc định nhúng trong `configs/config.go` (dùng khi không có env nào ghi đè). Cố tình không có
`allowed_origins`: CORS thuộc về gm-gateway, process duy nhất mà browser nói chuyện cùng.

## Docker

- `docker-compose.debug.yaml` + `Dockerfile.debug` — môi trường phát triển với debugger Delve (port 2302) và hot reload CompileDaemon; publish port `50060:50050` (gRPC) và `2302:2302` (Delve); mount repo vào `/app`, kèm `../gopkg` và `../gm-core`
- Admin port **không** được publish. gm-gateway đã chiếm host `50061`, và probe cũng vô dụng nếu gọi từ ngoài container. Dùng `make attach`, rồi curl `localhost:50051/metrics` từ bên trong.
