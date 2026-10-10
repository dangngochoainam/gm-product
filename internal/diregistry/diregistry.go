package diregistry

import (
	"github.com/dangngochoainam/gm-product/configs"
	"github.com/dangngochoainam/gm-product/internal/api"
	"github.com/dangngochoainam/gm-product/internal/repository/product"
	"github.com/dangngochoainam/gm-product/internal/usecase"
	"github.com/dangngochoainam/gopkg/copyhelper"
	"github.com/dangngochoainam/gopkg/gormhelper"
	"github.com/dangngochoainam/gopkg/httpclienthelper"
	"go.uber.org/dig"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	ConfigDIName string = "Config"
)

var (
	container *dig.Container
)

func BuildContainer() {
	ctn := dig.New()
	provide := func(constructor any) {
		if err := ctn.Provide(constructor); err != nil {
			panic(err)
		}
	}

	// CONFIG
	provide(configs.LoadConfig)

	// HELPER
	provide(copyhelper.NewModelConverter)
	provide(copyhelper.NewPbConverter)
	provide(func(cfg *configs.Config) httpclienthelper.HttpClient {
		hc := cfg.HttpClientConfig
		return httpclienthelper.NewHttpClientHelper(
			httpclienthelper.WithClientTimeout(hc.Timeout),
			httpclienthelper.WithRetryConfig(hc.MaxRetries, hc.RetryInitialDelay, hc.RetryMaxDelay, hc.RetryMultiplier),
			httpclienthelper.WithMaxResponseBodySize(hc.MaxResponseBodySize),
		)
	})

	// GORM
	provide(func(cfg *configs.Config) (*gorm.DB, error) {
		pg := cfg.DatabasePostgres
		opts := []gormhelper.GormHelperOption{}
		if pg.LoggingEnabled {
			opts = append(opts, gormhelper.WithLogLevel(logger.Info))
		}
		return gormhelper.NewGormHelper(
			&gormhelper.GormHelperConfig{
				Host:     pg.Host,
				Port:     pg.Port,
				Database: pg.Database,
				Schema:   pg.Schema,
				Username: pg.Username,
				Password: pg.Password,
			},
			opts...,
		)
	})

	// REPOSITORY
	provide(product.NewProductRepository)

	// USECASE
	provide(usecase.NewProductUseCase)

	// CONTROLLER
	provide(api.NewAPI)

	container = ctn
}

func GetDependency[T any]() T {
	var out T

	err := container.Invoke(func(dep T) {
		out = dep
	})
	if err != nil {
		panic(err)
	}

	return out
}
