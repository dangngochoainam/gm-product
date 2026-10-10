# gm-product

## 1. Resources

You only need two things installed locally:

- **Docker + Docker Compose**
- **`docker_stack` external network** with the following services running on it:

| Service | Container hostname | Port |
|---|---|---|
| PostgreSQL | `postgres` | 5432 |

Go tooling for running and debugging this service (CompileDaemon, Delve, Atlas) is pre-installed inside the container via `Dockerfile.debug` — no local setup required. Protobuf codegen is **not** here: it belongs to `gm-core` and runs on the host with `buf`.

If the network doesn't exist yet:
```bash
docker network create docker_stack
```

## 2. Run with watch

```bash
make start      # build image and start the container
make attach     # open a shell inside the container

# Inside the container:
make mod        # download and vendor deps (first time, or after go.mod changes)
make run-watch  # live reload via CompileDaemon + Delve debugger
```

Ports exposed to the host:

| Protocol | Host port |
|---|---|
| gRPC | `50060` |
| Delve debugger | `2302` |

This service serves **no public HTTP**. REST reaches it through gm-gateway over gRPC.
The admin listener on container port `50051` carries `/healthz`, `/readyz` and `/metrics`
only, and is not published to the host. See [`docs/service-ports.md`](docs/service-ports.md)
for why.

> First run only — apply DB migrations inside the container before starting the app:
> ```bash
> cd db/golang-migrate
> # edit export-env.local.sh with DB credentials (host: postgres, port: 5432)
> ./shmake build
> ./shmake diff init
> ./shmake apply
> ```

## 3. Update database schema (entity changes)

See [`db/golang-migrate/README.md`](db/golang-migrate/README.md) for the full workflow.

## 4. Regenerate proto

This repo has no `proto/` and no `pb/`. The product contract lives in the shared
[`gm-core`](../gm-core) module under `proto/product/v1/`, split into `dto/product.dto.proto`
(messages) and `service/product.service.proto` (rpcs). The generated code is imported here as
two packages:

```go
productdtov1 "github.com/dangngochoainam/gm-core/application/product/v1/dto"
productsvcv1 "github.com/dangngochoainam/gm-core/application/product/v1/service"
```

After editing either file:

```bash
cd ../gm-core && make gen-proto   # on the host — needs buf; this repo has no gen-proto target
make mod                          # inside the container — re-tidy and re-vendor
```

Commit the regenerated `gm-core/application/` alongside your `.proto` change, then update any
affected handlers in `internal/api/`. `gm-core` is bind-mounted at `/gm-core` in the debug container so
the `replace github.com/dangngochoainam/gm-core => ../gm-core` in `go.mod` resolves from `/app`.
