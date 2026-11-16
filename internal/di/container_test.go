package di_test

import (
	"pos/internal/adapters/memory"
	"pos/internal/application"
	"pos/internal/di"
	"pos/internal/domain"
	"testing"
)

type fakeStrategy struct{ calls int }

func (s *fakeStrategy) Calculate([]domain.Criterion, []domain.Evaluation) (float64, error) {
	s.calls++
	return 4.2, nil
}

type fakeExporter struct{ calls int }

func (e *fakeExporter) Export(string, []domain.Result) error { e.calls++; return nil }
func TestProvidersAreLazyAndDependenciesAreReplaced(t *testing.T) {
	strategy, exporter := &fakeStrategy{}, &fakeExporter{}
	created := 0
	c := di.New(di.Providers{
		Repository: func() application.Repository { created++; return memory.New() },
		Exporter:   func() application.ResultExporter { return exporter },
		Strategy:   func() domain.ScoringStrategy { return strategy },
	})
	if created != 0 {
		t.Fatal("фабрика вызвана до разрешения")
	}
	f := c.Facade()
	if f != c.Facade() || created != 1 {
		t.Fatal("экземпляр создаётся повторно")
	}
	if err := f.AddStudent("S", "Анна"); err != nil {
		t.Fatal(err)
	}
	if err := f.AddCriterion("C", "Критерий", 10, 1); err != nil {
		t.Fatal(err)
	}
	r, err := f.Total("S")
	if err != nil {
		t.Fatal(err)
	}
	if r.Score != 4.2 || strategy.calls != 1 {
		t.Fatal("стратегия не внедрена")
	}
	if err := f.Export("unused.csv"); err != nil {
		t.Fatal(err)
	}
	if exporter.calls != 1 {
		t.Fatal("экспорт не внедрён")
	}
}
func TestDefaultContainer(t *testing.T) {
	c := di.Default()
	if c.Facade() != c.Facade() {
		t.Fatal("разные экземпляры")
	}
	if err := c.Facade().AddStudent("S", "Анна"); err != nil {
		t.Fatal(err)
	}
	if err := c.Facade().AddCriterion("C", "Критерий", 10, 1); err != nil {
		t.Fatal(err)
	}
	r, err := c.Facade().Total("S")
	if err != nil || r.Score != 0 {
		t.Fatal(r, err)
	}
}
