package application

import (
	"fmt"
	"pos/internal/domain"
)

type Repository interface {
	Load() (*domain.Gradebook, error)
	Save(*domain.Gradebook) error
}

type ResultExporter interface {
	Export(string, []domain.Result) error
}

type Facade struct {
	repository Repository
	exporter   ResultExporter
	calculator domain.Calculator
}

func NewFacade(repository Repository, exporter ResultExporter, strategy domain.ScoringStrategy) *Facade {
	return &Facade{repository, exporter, domain.NewCalculator(strategy)}
}
func (f *Facade) change(action func(*domain.Gradebook) error) error {
	book, err := f.repository.Load()
	if err != nil {
		return err
	}
	if err := action(book); err != nil {
		return err
	}
	return f.repository.Save(book)
}
func (f *Facade) AddStudent(id, name string) error {
	return f.change(func(b *domain.Gradebook) error { return b.AddStudent(id, name) })
}
func (f *Facade) RenameStudent(id, name string) error {
	return f.change(func(b *domain.Gradebook) error { return b.RenameStudent(id, name) })
}
func (f *Facade) RemoveStudent(id string) error {
	return f.change(func(b *domain.Gradebook) error { return b.RemoveStudent(id) })
}
func (f *Facade) Students() ([]domain.Student, error) {
	b, err := f.repository.Load()
	if err != nil {
		return nil, err
	}
	return b.Students(), nil
}
func (f *Facade) AddCriterion(id, name string, maximum, weight float64) error {
	return f.change(func(b *domain.Gradebook) error { return b.AddCriterion(id, name, maximum, weight) })
}
func (f *Facade) UpdateCriterion(id, name string, maximum, weight float64) error {
	return f.change(func(b *domain.Gradebook) error { return b.UpdateCriterion(id, name, maximum, weight) })
}
func (f *Facade) RemoveCriterion(id string) error {
	return f.change(func(b *domain.Gradebook) error { return b.RemoveCriterion(id) })
}
func (f *Facade) Criteria() ([]domain.Criterion, error) {
	b, err := f.repository.Load()
	if err != nil {
		return nil, err
	}
	return b.Criteria(), nil
}
func (f *Facade) Grade(studentID, criterionID string, value float64) error {
	return f.change(func(b *domain.Gradebook) error { return b.Grade(studentID, criterionID, value) })
}
func (f *Facade) Cancel(studentID, criterionID string) error {
	return f.change(func(b *domain.Gradebook) error { return b.Cancel(studentID, criterionID) })
}
func (f *Facade) Evaluations(studentID string) ([]domain.Evaluation, error) {
	b, err := f.repository.Load()
	if err != nil {
		return nil, err
	}
	return b.Evaluations(studentID)
}
func (f *Facade) UsePolicy(name string) error {
	var strategy domain.ScoringStrategy
	switch name {
	case "zero":
		strategy = domain.MissingAsZero{}
	case "strict":
		strategy = domain.RequireComplete{}
	default:
		return fmt.Errorf("правило должно быть zero или strict")
	}
	f.calculator = domain.NewCalculator(strategy)
	return nil
}
func (f *Facade) Total(studentID string) (domain.Result, error) {
	b, err := f.repository.Load()
	if err != nil {
		return domain.Result{}, err
	}
	return f.calculator.Calculate(b, studentID)
}

type Report struct {
	Results                   []domain.Result
	Average, Minimum, Maximum float64
}

func (f *Facade) Totals() (Report, error) {
	b, err := f.repository.Load()
	if err != nil {
		return Report{}, err
	}
	if len(b.Criteria()) == 0 {
		return Report{}, fmt.Errorf("рубрикатор пуст: добавьте критерии")
	}
	report := Report{Results: make([]domain.Result, 0)}
	for _, s := range b.Students() {
		result, err := f.calculator.Calculate(b, s.ID())
		if err != nil {
			return Report{}, err
		}
		if len(report.Results) == 0 || result.Score < report.Minimum {
			report.Minimum = result.Score
		}
		if result.Score > report.Maximum {
			report.Maximum = result.Score
		}
		report.Average += result.Score
		report.Results = append(report.Results, result)
	}
	if len(report.Results) != 0 {
		report.Average /= float64(len(report.Results))
	}
	return report, nil
}
func (f *Facade) Export(path string) error {
	report, err := f.Totals()
	if err != nil {
		return err
	}
	return f.exporter.Export(path, report.Results)
}
