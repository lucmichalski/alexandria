// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package dep

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/tracer"
	"github.com/go-kit/kit/log"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/dependency"
	"github.com/maestre3d/alexandria/author-service/pkg/author"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/maestre3d/alexandria/author-service/pkg/service"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/bind"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/proxy"
)

// Injectors from wire.go:

func InjectTransportService() (*service.Transport, func(), error) {
	logLogger := logger.NewZapLogger()
	authorInteractor, cleanup, err := provideAuthorInteractor(logLogger)
	if err != nil {
		return nil, nil, err
	}
	context := provideContext()
	kernel, err := config.NewKernel(context)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	zipkinTracer, cleanup2 := tracer.NewZipkin(kernel)
	opentracingTracer := tracer.WrapZipkinOpenTracing(kernel, zipkinTracer)
	authorServer := bind.NewAuthorRPC(authorInteractor, logLogger, opentracingTracer, zipkinTracer)
	servers := provideRPCServers(authorServer)
	server, cleanup3 := proxy.NewRPC(servers)
	authorHandler := bind.NewAuthorHTTP(authorInteractor, logLogger, opentracingTracer, zipkinTracer)
	v := provideHTTPHandlers(authorHandler)
	http, cleanup4 := proxy.NewHTTP(kernel, v...)
	transport := service.NewTransport(server, http, kernel)
	return transport, func() {
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}

// wire.go:

var authorInteractorSet = wire.NewSet(logger.NewZapLogger, provideAuthorInteractor)

var httpProxySet = wire.NewSet(
	authorInteractorSet,
	provideContext, config.NewKernel, tracer.NewZipkin, tracer.WrapZipkinOpenTracing, bind.NewAuthorHTTP, provideHTTPHandlers, proxy.NewHTTP,
)

var rpcProxySet = wire.NewSet(bind.NewAuthorRPC, provideRPCServers, proxy.NewRPC)

func provideContext() context.Context {
	return context.Background()
}

func provideAuthorInteractor(logger2 log.Logger) (usecase.AuthorInteractor, func(), error) {
	authorUseCase, cleanup, err := dependency.InjectAuthorUseCase()

	authorService := author.WrapAuthorInstrumentation(authorUseCase, logger2)

	return authorService, cleanup, err
}

// Bind/Map used http handlers
func provideHTTPHandlers(authorHandler *bind.AuthorHandler) []proxy.Handler {
	handlers := make([]proxy.Handler, 0)
	handlers = append(handlers, authorHandler)
	return handlers
}

// Bind/Map used rpc actions
func provideRPCServers(authorHandler pb.AuthorServer) *proxy.Servers {
	return &proxy.Servers{authorHandler}
}