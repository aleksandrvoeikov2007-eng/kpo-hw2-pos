package di

import (
	"pos/internal/adapters/csvexport"
	"pos/internal/adapters/memory"
	"pos/internal/application"
	"pos/internal/domain"
)

type Providers struct {
	Repository func() application.Repository
	Exporter   func() application.ResultExporter
	Strategy   func() domain.ScoringStrategy
}

type Container struct {
	providers Providers
	facade    *application.Facade
}

func New(providers Providers) *Container { return &Container{providers: providers} }
func Default() *Container {
	return New(Providers{
		Repository: func() application.Repository { return memory.New() },
		Exporter:   func() application.ResultExporter { return csvexport.Exporter{} },
		Strategy:   func() domain.ScoringStrategy { return domain.MissingAsZero{} },
	})
}
func (c *Container) Facade() *application.Facade {
	if c.facade == nil {
		c.facade = application.NewFacade(c.providers.Repository(), c.providers.Exporter(), c.providers.Strategy())
	}
	return c.facade
}
