// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package dep

import (
	"context"
	zipkin2 "contrib.go.opencensus.io/exporter/zipkin"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/tracer"
	"github.com/alexandria-oss/core/transport"
	"github.com/alexandria-oss/core/transport/proxy"
	"github.com/go-kit/kit/log"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/blob-service/internal/dependency"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	"github.com/maestre3d/alexandria/blob-service/pkg/transport/bind"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/reporter"
	"github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

// Injectors from wire.go:

func InjectTransportService() (*transport.Transport, func(), error) {
	v := provideRPCServers()
	server, cleanup := proxy.NewRPC(v)
	context := provideContext()
	kernel, err := config.NewKernel(context)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	logLogger := logger.NewZapLogger()
	blobInteractor, cleanup2, err := provideBlobInteractor(logLogger)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	reporter, cleanup3 := provideZipkinReporter(kernel)
	endpoint := provideZipkinEndpoint(kernel)
	zipkinTracer := provideZipkinTracer(kernel, reporter, endpoint)
	opentracingTracer := tracer.WrapZipkinOpenTracing(kernel, zipkinTracer)
	blobHandler := bind.NewBlobHandler(blobInteractor, logLogger, opentracingTracer, zipkinTracer)
	v2 := provideHTTPHandlers(blobHandler)
	http, cleanup4 := proxy.NewHTTP(kernel, v2...)
	blobSagaInteractor, cleanup5, err := provideBlobSagaInteractor(logLogger)
	if err != nil {
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	blobEventConsumer := bind.NewBlobEventConsumer(blobSagaInteractor, logLogger, kernel)
	v3 := provideEventConsumers(blobEventConsumer)
	event, cleanup6, err := proxy.NewEvent(context, kernel, v3...)
	if err != nil {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	transportTransport := transport.NewTransport(server, http, event, kernel)
	return transportTransport, func() {
		cleanup6()
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}

// wire.go:

var Ctx = context.Background()

var interactorSet = wire.NewSet(
	provideContext, logger.NewZapLogger, provideBlobInteractor,
	provideBlobSagaInteractor,
)

var zipkinSet = wire.NewSet(
	provideZipkinReporter,
	provideZipkinEndpoint,
	provideZipkinTracer,
)

var httpProxySet = wire.NewSet(
	interactorSet, config.NewKernel, zipkinSet, tracer.WrapZipkinOpenTracing, bind.NewBlobHandler, provideHTTPHandlers, proxy.NewHTTP,
)

var eventProxySet = wire.NewSet(bind.NewBlobEventConsumer, provideEventConsumers, proxy.NewEvent)

func provideContext() context.Context {
	return Ctx
}

func provideBlobInteractor(log2 log.Logger) (usecase.BlobInteractor, func(), error) {
	dependency.Ctx = Ctx

	interactor, cleanup, err := dependency.InjectBlobUseCase()
	svc := blob.WrapBlobInstrumentation(interactor, log2)

	return svc, cleanup, err
}

func provideBlobSagaInteractor(log2 log.Logger) (usecase.BlobSagaInteractor, func(), error) {
	dependency.Ctx = Ctx

	interactor, cleanup, err := dependency.InjectBlobSagaUseCase()
	svc := blob.WrapBlobSagaInstrumentation(interactor, log2)

	return svc, cleanup, err
}

// Bind/Map used http handlers
func provideHTTPHandlers(blobHandler *bind.BlobHandler) []proxy.Handler {
	handlers := make([]proxy.Handler, 0)
	handlers = append(handlers, blobHandler)
	return handlers
}

// Bind/Map used rpc servers
func provideRPCServers() []proxy.RPCServer {
	servers := make([]proxy.RPCServer, 0)

	return servers
}

// Bind/Map used event consumers
func provideEventConsumers(blobConsumer *bind.BlobEventConsumer) []proxy.Consumer {
	consumers := make([]proxy.Consumer, 0)
	consumers = append(consumers, blobConsumer)
	return consumers
}

// NewZipkin returns a zipkin tracing consumer
func provideZipkinReporter(cfg *config.Kernel) (reporter.Reporter, func()) {
	if cfg.Tracing.ZipkinHost != "" && cfg.Tracing.ZipkinEndpoint != "" {
		zipkinReporter := http.NewReporter(cfg.Tracing.ZipkinHost)
		cleanup := func() {
			_ = zipkinReporter.Close()
		}

		return zipkinReporter, cleanup
	}

	return nil, nil
}

// NewZipkin returns a zipkin tracing consumer
func provideZipkinEndpoint(cfg *config.Kernel) *model.Endpoint {
	if cfg.Tracing.ZipkinHost != "" && cfg.Tracing.ZipkinEndpoint != "" {
		zipkinEndpoint, err := zipkin.NewEndpoint(cfg.Service, cfg.Tracing.ZipkinEndpoint)
		if err != nil {
			return nil
		}

		return zipkinEndpoint
	}

	return nil
}

// NewZipkin returns a zipkin tracing consumer
func provideZipkinTracer(cfg *config.Kernel, r reporter.Reporter, ep *model.Endpoint) *zipkin.Tracer {
	if cfg.Tracing.ZipkinHost != "" && cfg.Tracing.ZipkinEndpoint != "" {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
		trace.RegisterExporter(zipkin2.NewExporter(r, ep))

		zipkinTrace, err := zipkin.NewTracer(r, zipkin.WithLocalEndpoint(ep))
		if err != nil {
			return nil
		}

		return zipkinTrace
	}

	return nil
}
