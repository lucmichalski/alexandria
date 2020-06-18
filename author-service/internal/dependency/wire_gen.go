// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package dependency

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/author-service/internal/interactor"
)

// Injectors from wire.go:

func InjectAuthorUseCase() (*interactor.AuthorUseCase, func(), error) {
	logLogger := logger.NewZapLogger()
	context := provideContext()
	kernel, err := config.NewKernel(context)
	if err != nil {
		return nil, nil, err
	}
	db, cleanup, err := persistence.NewPostgresPool(context, kernel)
	if err != nil {
		return nil, nil, err
	}
	client, cleanup2, err := persistence.NewRedisPool(kernel)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	authorPostgresRepository := infrastructure.NewAuthorPostgresRepository(db, client, logLogger)
	authorKafkaEventBus := infrastructure.NewAuthorKafkaEventBus(kernel)
	authorUseCase := interactor.NewAuthorUseCase(logLogger, authorPostgresRepository, authorKafkaEventBus)
	return authorUseCase, func() {
		cleanup2()
		cleanup()
	}, nil
}

// wire.go:

var Ctx context.Context = context.Background()

var configSet = wire.NewSet(
	provideContext, config.NewKernel,
)

var dBMSPoolSet = wire.NewSet(
	configSet, persistence.NewPostgresPool,
)

var authorDBMSRepositorySet = wire.NewSet(
	dBMSPoolSet, logger.NewZapLogger, persistence.NewRedisPool, wire.Bind(new(domain.AuthorRepository), new(*infrastructure.AuthorPostgresRepository)), infrastructure.NewAuthorPostgresRepository,
)

func provideContext() context.Context {
	return Ctx
}
