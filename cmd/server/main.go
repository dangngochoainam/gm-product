package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	productsvcv1 "github.com/dangngochoainam/gm-core/application/product/v1/service"
	"github.com/dangngochoainam/gm-product/configs"
	"github.com/dangngochoainam/gm-product/internal/diregistry"
	"github.com/dangngochoainam/gm-product/internal/health"
	"github.com/dangngochoainam/gopkg/commonhelper"
	"github.com/dangngochoainam/gopkg/loghelper"
	"github.com/dangngochoainam/gopkg/prometheushelper"
	"github.com/dangngochoainam/gopkg/serverhelper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

// gm-product serves gRPC only. All external HTTP terminates at gm-gateway, which
// translates REST to gRPC from the google.api.http annotations in gm-core. This
// process therefore runs no grpc-gateway mux and no CORS: a second door into the
// same handlers would bypass every gateway middleware, including the header
// allowlist that stops a client from forging its own trace ID.
//
// The one HTTP listener left is the admin port. It carries probes and metrics
// only, is never routed through an ingress, and is documented in
// docs/service-ports.md.
func main() {

	diregistry.BuildContainer()

	config := diregistry.GetDependency[*configs.Config]()

	api := diregistry.GetDependency[productsvcv1.ProductServiceServer]()

	err := loghelper.InitZapLogger(config.Env, "gm-product")
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer func() { _ = loghelper.Logger.Sync() }()

	ctx := context.Background()
	ctx = context.WithValue(ctx, commonhelper.TRACE_ID, "123123")
	loghelper.Logger.WithContext(ctx).Info("Init logger success")

	// Cancelled on SIGINT/SIGTERM, which starts the drain below.
	shutdownCtx, stopListeningForSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopListeningForSignals()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// The registry has to reach the interceptors, so it is built before the
	// server and handed to the same instance the admin listener exposes.
	registry := prometheushelper.GetRegistry()

	grpcServer := serverhelper.CreateGrpcServer(serverhelper.WithPrometheusRegistry(registry))
	productsvcv1.RegisterProductServiceServer(grpcServer, api)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	loghelper.Logger.Infow("grpc server running", "address", lis.Addr().String())

	readiness := health.NewReadiness()
	adminServer, adminListener := startAdminServer(config, registry, readiness)

	// The gRPC listener is up, so the pod may take traffic.
	readiness.SetReady()
	loghelper.Logger.Infow("admin server running", "address", adminListener.Addr().String())

	<-shutdownCtx.Done()
	drainAndShutdown(grpcServer, adminServer, readiness, config)
}

// startAdminServer exposes probes and metrics. Nothing here is business traffic,
// so it deliberately shares no middleware, no auth and no CORS with the gRPC
// path. Keeping it on its own port is what lets the Kubernetes Service publish
// the gRPC port alone.
func startAdminServer(config *configs.Config, registry prometheus.Gatherer, readiness *health.Readiness) (*http.Server, net.Listener) {
	mux := http.NewServeMux()
	health.Register(mux, readiness)
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.AdminPort))
	if err != nil {
		log.Panic("Admin server: can not create listener - err: ", err)
	}

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		err := server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("admin server: %v", err)
		}
	}()

	return server, listener
}

// drainAndShutdown fails readiness first, then waits before stopping anything.
// Kubernetes removes the pod from the Service endpoint list asynchronously, so
// closing the listener immediately on SIGTERM drops calls that are still being
// routed here. gRPC stops before the admin listener so probes keep answering
// for the whole drain. Keep terminationGracePeriodSeconds above
// shutdown_drain_delay + shutdown_timeout.
func drainAndShutdown(grpcServer *grpc.Server, adminServer *http.Server, readiness *health.Readiness, config *configs.Config) {
	readiness.SetNotReady()
	loghelper.Logger.Infow("shutdown signal received, draining", "drain_delay", config.ShutdownDrainDelay.String())
	time.Sleep(config.ShutdownDrainDelay)

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(config.ShutdownTimeout):
		loghelper.Logger.Warn("graceful stop timed out, forcing gRPC server down")
		grpcServer.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()
	if err := adminServer.Shutdown(ctx); err != nil {
		loghelper.Logger.Errorw("admin server shutdown failed", "error", err)
		return
	}
	loghelper.Logger.Info("graceful shutdown complete")
}
