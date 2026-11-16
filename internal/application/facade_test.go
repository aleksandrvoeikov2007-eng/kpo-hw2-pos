package application_test

import (
	"errors"
	"math"
	"pos/internal/adapters/memory"
	"pos/internal/application"
	"pos/internal/domain"
	"testing"
)

type exporter struct {
	calls   int
	path    string
	results []domain.Result
	err     error
}

func (e *exporter) Export(path string, results []domain.Result) error {
	e.calls++
	e.path = path
	e.results = results
	return e.err
}
func ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func setup(t *testing.T) (*application.Facade, *exporter) {
	t.Helper()
	e := &exporter{}
	f := application.NewFacade(memory.New(), e, domain.MissingAsZero{})
	ok(t, f.AddStudent("S1", "Анна"))
	ok(t, f.AddStudent("S2", "Борис"))
	ok(t, f.AddCriterion("C1", "Практика", 10, 0.4))
	ok(t, f.AddCriterion("C2", "Экзамен", 20, 0.6))
	ok(t, f.Grade("S1", "C1", 8))
	ok(t, f.Grade("S1", "C2", 15))
	return f, e
}
func TestFacadeReportAndExport(t *testing.T) {
	f, e := setup(t)
	report, err := f.Totals()
	ok(t, err)
	if len(report.Results) != 2 || math.Abs(report.Average-3.85) > 1e-9 || report.Minimum != 0 || math.Abs(report.Maximum-7.7) > 1e-9 {
		t.Fatal(report)
	}
	ok(t, f.Export("results.csv"))
	if e.calls != 1 || e.path != "results.csv" || len(e.results) != 2 {
		t.Fatal(e)
	}
	e.err = errors.New("экспорт недоступен")
	if err := f.Export("results.csv"); !errors.Is(err, e.err) {
		t.Fatal(err)
	}
	ok(t, f.UsePolicy("strict"))
	if _, err := f.Totals(); err == nil {
		t.Fatal("неполные оценки приняты")
	}
	if err := f.Export("bad.csv"); err == nil || e.calls != 2 {
		t.Fatal("экспорт неполного отчёта")
	}
	if err := f.UsePolicy("unknown"); err == nil {
		t.Fatal("неизвестная стратегия принята")
	}
	if _, err := f.Total("S2"); err == nil {
		t.Fatal("ошибка выбора изменила стратегию")
	}
	ok(t, f.UsePolicy("zero"))
	ok(t, f.Grade("S2", "C1", 10))
	ok(t, f.Grade("S2", "C2", 20))
	ok(t, f.RenameStudent("S2", "Борис Сергеев"))
	ok(t, f.UpdateCriterion("C2", "Экзамен", 20, 0.2))
	ok(t, f.Cancel("S1", "C2"))
	list, err := f.Evaluations("S1")
	ok(t, err)
	if len(list) != 1 {
		t.Fatal(list)
	}
	ok(t, f.Cancel("S2", "C2"))
	ok(t, f.RemoveCriterion("C2"))
	ok(t, f.RemoveStudent("S2"))
	students, err := f.Students()
	ok(t, err)
	criteria, err := f.Criteria()
	ok(t, err)
	if len(students) != 1 || len(criteria) != 1 {
		t.Fatal(students, criteria)
	}
	r, err := f.Total("S1")
	ok(t, err)
	if math.Abs(r.Score-8) > 1e-9 {
		t.Fatal(r)
	}
}
func TestEmptyReport(t *testing.T) {
	f := application.NewFacade(memory.New(), &exporter{}, domain.MissingAsZero{})
	if _, err := f.Totals(); err == nil {
		t.Fatal("пустой рубрикатор принят")
	}
	ok(t, f.AddCriterion("C", "Критерий", 10, 1))
	report, err := f.Totals()
	ok(t, err)
	if len(report.Results) != 0 || report.Average != 0 {
		t.Fatal(report)
	}
}

type failingRepository struct {
	base             *memory.Repository
	loadErr, saveErr error
	saves            int
}

func (r *failingRepository) Load() (*domain.Gradebook, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.base.Load()
}
func (r *failingRepository) Save(b *domain.Gradebook) error {
	r.saves++
	if r.saveErr != nil {
		return r.saveErr
	}
	return r.base.Save(b)
}
func TestRepositoryFailureAndFailedMutation(t *testing.T) {
	failure := errors.New("хранилище недоступно")
	repo := &failingRepository{base: memory.New(), loadErr: failure}
	f := application.NewFacade(repo, &exporter{}, domain.MissingAsZero{})
	if err := f.AddStudent("S", "Имя"); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	reads := []func() error{
		func() error { _, err := f.Students(); return err },
		func() error { _, err := f.Criteria(); return err },
		func() error { _, err := f.Evaluations("S"); return err },
		func() error { _, err := f.Total("S"); return err },
		func() error { _, err := f.Totals(); return err },
	}
	for _, read := range reads {
		if !errors.Is(read(), failure) {
			t.Fatal("ошибка хранилища потеряна")
		}
	}
	repo.loadErr = nil
	repo.saveErr = failure
	if err := f.AddStudent("S", "Имя"); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	list, err := f.Students()
	ok(t, err)
	if len(list) != 0 {
		t.Fatal("неудачное сохранение изменило состояние")
	}
	repo.saveErr = nil
	ok(t, f.AddStudent("S", "Имя"))
	calls := repo.saves
	if err := f.AddStudent("S", "Повтор"); err == nil || repo.saves != calls {
		t.Fatal("неудачное действие вызывает сохранение")
	}
}
