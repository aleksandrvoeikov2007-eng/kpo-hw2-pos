package domain

import "fmt"

type ScoringStrategy interface {
	Calculate([]Criterion, []Evaluation) (float64, error)
}

type MissingAsZero struct{}

func (MissingAsZero) Calculate(criteria []Criterion, evaluations []Evaluation) (float64, error) {
	return weighted(criteria, evaluations, false)
}

type RequireComplete struct{}

func (RequireComplete) Calculate(criteria []Criterion, evaluations []Evaluation) (float64, error) {
	return weighted(criteria, evaluations, true)
}

func weighted(criteria []Criterion, evaluations []Evaluation, strict bool) (float64, error) {
	if len(criteria) == 0 {
		return 0, fmt.Errorf("рубрикатор пуст: добавьте критерии")
	}
	scores := make(map[string]float64, len(evaluations))
	for _, e := range evaluations {
		scores[e.CriterionID()] = e.Score().Value()
	}
	var sum, weights float64
	for _, c := range criteria {
		value, exists := scores[c.ID()]
		if strict && !exists {
			return 0, fmt.Errorf("не выставлена оценка по критерию %q", c.ID())
		}
		sum += value / c.Maximum() * c.Weight()
		weights += c.Weight()
	}
	return 10 * (sum / weights), nil
}

type Result struct {
	StudentID   string
	StudentName string
	Score       float64
	Graded      int
	Total       int
}

type Calculator struct{ strategy ScoringStrategy }

func NewCalculator(strategy ScoringStrategy) Calculator { return Calculator{strategy} }
func (c Calculator) Calculate(book *Gradebook, studentID string) (Result, error) {
	student, err := book.Student(studentID)
	if err != nil {
		return Result{}, err
	}
	criteria := book.Criteria()
	evaluations, err := book.Evaluations(studentID)
	if err != nil {
		return Result{}, err
	}
	score, err := c.strategy.Calculate(criteria, evaluations)
	if err != nil {
		return Result{}, fmt.Errorf("студент %s: %w", studentID, err)
	}
	return Result{student.ID(), student.Name(), score, len(evaluations), len(criteria)}, nil
}
