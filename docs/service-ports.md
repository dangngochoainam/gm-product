# Vì sao gm-product không phục vụ HTTP public

Bản ghi quyết định cho thay đổi đã gỡ grpc-gateway mux in-process khỏi service này.
Viết ngày 2026-09-12.

## Đã thay đổi gì

Trước đây, `cmd/server/main.go` khởi động một gRPC server, tự dial **chính nó** qua
loopback, register một grpc-gateway mux lên connection đó, bọc mux trong CORS rồi
phục vụ kết quả trên port HTTP. Đó là hình dạng mà `codebase-go-grpc` mang theo, và
nó đúng cho một service không có gateway đứng trước.

Bây giờ process phục vụ gRPC cho traffic nghiệp vụ và một admin listener cho probe
với metrics. Không mux, không CORS, không self-dial.

| Port | Phục vụ gì | Ai gọi |
|---|---|---|
| gRPC | mọi rpc trong `product.v1.service.ProductService` | gm-gateway, và các service khác trong cluster |
| admin HTTP | `/healthz`, `/readyz`, `/metrics` | kubelet và Prometheus. Không bao giờ là đích của ingress |

## Vì sao port HTTP public phải biến mất

**Hai cửa, một bộ handler, hai tư thế bảo mật.** Mọi request đi qua gm-gateway đều
chạy qua RequestID, AccessLog, Recovery, StripIdentityHeaders, MaxBodyBytes và
Timeout. Request đến thẳng port HTTP của service này thì không qua cái nào. Gateway
chỉ còn là một gợi ý, không phải một ranh giới.

**Trace ID có thể bị giả mạo.** `grpcmiddlewarehelper.SetIDUnaryServerInterceptor`
lấy correlation ID thẳng từ gRPC metadata đến và tin luôn. Thứ ngăn client tự bịa ra
một cái là allowlist header incoming ở gateway, vốn từ chối mọi thứ ngoài một tập cố
định và không bao giờ cho tiền tố `grpc-metadata-` đi qua. Gọi thẳng vào service này
là bỏ qua đúng bước kiểm tra đó. Hôm nay chuyện đó mới chỉ làm hỏng log; ngay khi một
header identity đi vào cùng đường dẫn ấy thì nó là một lỗ hổng thật.

**Một contract, register hai lần.** Cùng bộ annotation `google.api.http` được mount ở
đây và ở gm-gateway. Không có gì ép hai process phải build trên cùng một version của
gm-core, nên route có thể lệch nhau trong im lặng.

**Một hop loopback cho mỗi request.** Mỗi lời gọi HTTP phải vượt một TCP connection
tới chính gRPC server của process rồi trả giá thêm một vòng protobuf nữa, cho một
đường đi vốn đã có cửa trước tốt hơn.

**CORS trong một backend service chỉ là đồ trang trí.** Không browser nào chạm tới
process này. Allowlist origin thuộc về gm-gateway, và `allowed_origins` đã bị bỏ khỏi
config của service này vì lý do đó.

**Chi phí vận hành.** Một port được publish là một Service port, một rule
NetworkPolicy và một listener phải drain khi nhận SIGTERM. Nó chẳng mua lại được gì
cho những chi phí đó.

Thứ duy nhất mất đi là khả năng `curl` thẳng vào service mà không qua gateway.
`grpcurl`, hoặc `kubectl port-forward` vào port gRPC, thay thế được.

## Vì sao vẫn giữ một admin listener

Bỏ HTTP hoàn toàn sẽ mất ba thứ thực sự cần đến nó:

- **Prometheus scrape qua HTTP.** `gopkg` đã có sẵn các mảnh ghép:
  `prometheushelper.GetRegistry` cùng option `WithPrometheusRegistry` trên
  `serverhelper.CreateGrpcServer`, thứ cài đặt các interceptor metric của gRPC.
  Service này trước giờ chưa bao giờ truyền registry vào và chưa bao giờ expose
  handler, nên đến giờ chưa có metric nào rời khỏi process. Cả hai đầu đã được nối
  trong `main.go`.
- **Probe.** Kubernetes hỗ trợ gRPC probe native từ 1.24, nên hoàn toàn có thể làm
  trên port gRPC. Chọn HTTP vì gm-gateway đã có sẵn pattern `/healthz` và `/readyz`
  trong `internal/health`, và một hình dạng chung cho cả fleet thì hơn hai.
- **pprof**, cho lúc cần soi một vụ rò goroutine. Chưa mount; nhưng listener này là
  chỗ nó sẽ nằm.

Giữ những thứ này trên một port riêng chính là điều cho phép Kubernetes Service chỉ
publish mỗi port gRPC. Admin port ở lại bên trong cluster.

## Thứ tự shutdown

`readyz` fail trước, rồi process chờ `shutdown_drain_delay` trước khi dừng bất cứ thứ
gì, vì Kubernetes xóa pod khỏi endpoint list một cách bất đồng bộ và request vẫn tiếp
tục đổ về sau SIGTERM. gRPC dừng kế tiếp, admin listener dừng cuối cùng, để probe còn
trả lời suốt quá trình drain. Giữ `terminationGracePeriodSeconds` lớn hơn
`shutdown_drain_delay + shutdown_timeout`.

Đây là lý do `serverhelper.HandleGracefulShutdownServer` không còn được dùng ở đây:
nó dừng HTTP trước gRPC và không có hook readiness, thứ tự đó là sai khi probe nằm ở
phía HTTP.

## Điều này có ý nghĩa gì với các domain service khác

gm-order, gm-payment và gm-inventory thừa hưởng đúng quy tắc này khi được derive từ
`codebase-go-grpc`: xóa mux, xóa CORS wrapper và xóa self-dial ngay ở bước đầu tiên,
giữ lại admin listener. gm-gateway là ngoại lệ, vì HTTP public chính là toàn bộ công
việc của nó.
