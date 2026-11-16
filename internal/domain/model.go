package domain

import (
	"fmt"
	"math"
	"strings"
)

func finite(n float64) bool { return !math.IsNaN(n) && !math.IsInf(n, 0) }
func identity(id, name string) (string, string, error) {
	id, name = strings.TrimSpace(id), strings.TrimSpace(name)
	if id == "" || name == "" {
		return "", "", fmt.Errorf("идентификатор и название обязательны")
	}
	return id, name, nil
}

type Student struct{ id, name string }

func NewStudent(id, name string) (Student, error) {
	id, name, err := identity(id, name)
	return Student{id, name}, err
}
func (s Student) ID() string   { return s.id }
func (s Student) Name() string { return s.name }

type Criterion struct {
	id, name        string
	maximum, weight float64
}

func NewCriterion(id, name string, maximum, weight float64) (Criterion, error) {
	id, name, err := identity(id, name)
	if err != nil {
		return Criterion{}, err
	}
	if !finite(maximum) || maximum <= 0 {
		return Criterion{}, fmt.Errorf("максимум должен быть конечным положительным числом")
	}
	if !finite(weight) || weight <= 0 || weight > 1 {
		return Criterion{}, fmt.Errorf("вес должен быть больше 0 и не больше 1")
	}
	return Criterion{id, name, maximum, weight}, nil
}
func (c Criterion) ID() string       { return c.id }
func (c Criterion) Name() string     { return c.name }
func (c Criterion) Maximum() float64 { return c.maximum }
func (c Criterion) Weight() float64  { return c.weight }

type Score struct{ value float64 }

func NewScore(value, maximum float64) (Score, error) {
	if !finite(maximum) || maximum <= 0 || !finite(value) || value < 0 || value > maximum {
		return Score{}, fmt.Errorf("балл должен быть в диапазоне от 0 до %g", maximum)
	}
	return Score{value}, nil
}
func (s Score) Value() float64 { return s.value }

type Evaluation struct {
	criterionID string
	score       Score
}

func (e Evaluation) CriterionID() string { return e.criterionID }
func (e Evaluation) Score() Score        { return e.score }

type EvaluationFactory struct{}

func (EvaluationFactory) Create(rubric []Criterion, criterionID string, value float64) (Evaluation, error) {
	for _, c := range rubric {
		if c.ID() == criterionID {
			score, err := NewScore(value, c.Maximum())
			if err != nil {
				return Evaluation{}, err
			}
			return Evaluation{criterionID, score}, nil
		}
	}
	return Evaluation{}, fmt.Errorf("критерий %q отсутствует в рубрикаторе", criterionID)
}
