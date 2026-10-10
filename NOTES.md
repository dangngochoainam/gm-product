# Những việc đang hoãn

## Lỗi lint có sẵn

`.golangci.yml` đặt `new-from-merge-base: origin/master`, nên các lỗi dưới đây (có từ trước khi thêm lint) không làm `make verify` fail. Sửa xong hết thì bỏ dòng `new-from-merge-base` để lint chặn toàn repo.

Kết quả golangci-lint v2.14.0, linter `standard` + gofmt + goimports:

- `cmd/client/main.go:32` — errcheck: `conn.Close`
- `cmd/server/main.go:47` — errcheck: `defer loghelper.Logger.Sync()`
- `internal/diregistry/diregistry.go:28`, `:31`, `:32` — errcheck: `ctn.Provide`
