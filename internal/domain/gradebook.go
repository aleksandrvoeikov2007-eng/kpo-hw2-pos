package domain

import (
	"fmt"
	"sort"
)

type Gradebook struct {
	students    map[string]Student
	criteria    map[string]Criterion
	evaluations map[string]map[string]Evaluation
}

func NewGradebook() *Gradebook {
	return &Gradebook{make(map[string]Student), make(map[string]Criterion), make(map[string]map[string]Evaluation)}
}
func (b *Gradebook) AddStudent(id, name string) error {
	s, err := NewStudent(id, name)
	if err != nil {
		return err
	}
	if _, exists := b.students[s.ID()]; exists {
		return fmt.Errorf("студент %q уже существует", s.ID())
	}
	b.students[s.ID()] = s
	return nil
}
func (b *Gradebook) Student(id string) (Student, error) {
	s, ok := b.students[id]
	if !ok {
		return Student{}, fmt.Errorf("студент %q не найден", id)
	}
	return s, nil
}
func (b *Gradebook) RenameStudent(id, name string) error {
	if _, err := b.Student(id); err != nil {
		return err
	}
	s, err := NewStudent(id, name)
	if err != nil {
		return err
	}
	b.students[id] = s
	return nil
}
func (b *Gradebook) RemoveStudent(id string) error {
	if _, err := b.Student(id); err != nil {
		return err
	}
	delete(b.students, id)
	delete(b.evaluations, id)
	return nil
}
func (b *Gradebook) AddCriterion(id, name string, maximum, weight float64) error {
	c, err := NewCriterion(id, name, maximum, weight)
	if err != nil {
		return err
	}
	if _, exists := b.criteria[c.ID()]; exists {
		return fmt.Errorf("критерий %q уже существует", c.ID())
	}
	b.criteria[c.ID()] = c
	return nil
}
func (b *Gradebook) UpdateCriterion(id, name string, maximum, weight float64) error {
	if _, exists := b.criteria[id]; !exists {
		return fmt.Errorf("критерий %q не найден", id)
	}
	c, err := NewCriterion(id, name, maximum, weight)
	if err != nil {
		return err
	}
	for _, list := range b.evaluations {
		if e, ok := list[id]; ok && e.Score().Value() > maximum {
			return fmt.Errorf("новый максимум меньше уже выставленной оценки")
		}
	}
	b.criteria[id] = c
	return nil
}
func (b *Gradebook) RemoveCriterion(id string) error {
	if _, exists := b.criteria[id]; !exists {
		return fmt.Errorf("критерий %q не найден", id)
	}
	for _, list := range b.evaluations {
		if _, ok := list[id]; ok {
			return fmt.Errorf("сначала отмените оценки по этому критерию")
		}
	}
	delete(b.criteria, id)
	return nil
}
func (b *Gradebook) Grade(studentID, criterionID string, value float64) error {
	if _, err := b.Student(studentID); err != nil {
		return err
	}
	e, err := (EvaluationFactory{}).Create(b.Criteria(), criterionID, value)
	if err != nil {
		return err
	}
	if b.evaluations[studentID] == nil {
		b.evaluations[studentID] = make(map[string]Evaluation)
	}
	b.evaluations[studentID][criterionID] = e
	return nil
}
func (b *Gradebook) Cancel(studentID, criterionID string) error {
	if _, err := b.Student(studentID); err != nil {
		return err
	}
	if _, exists := b.evaluations[studentID][criterionID]; !exists {
		return fmt.Errorf("оценка по критерию %q не найдена", criterionID)
	}
	delete(b.evaluations[studentID], criterionID)
	return nil
}
func (b *Gradebook) Students() []Student {
	result := make([]Student, 0, len(b.students))
	for _, s := range b.students {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID() < result[j].ID() })
	return result
}
func (b *Gradebook) Criteria() []Criterion {
	result := make([]Criterion, 0, len(b.criteria))
	for _, c := range b.criteria {
		result = append(result, c)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID() < result[j].ID() })
	return result
}
func (b *Gradebook) Evaluations(studentID string) ([]Evaluation, error) {
	if _, err := b.Student(studentID); err != nil {
		return nil, err
	}
	result := make([]Evaluation, 0, len(b.evaluations[studentID]))
	for _, e := range b.evaluations[studentID] {
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CriterionID() < result[j].CriterionID() })
	return result, nil
}
func (b *Gradebook) Clone() *Gradebook {
	result := NewGradebook()
	for id, s := range b.students {
		result.students[id] = s
	}
	for id, c := range b.criteria {
		result.criteria[id] = c
	}
	for id, list := range b.evaluations {
		result.evaluations[id] = make(map[string]Evaluation)
		for cid, e := range list {
			result.evaluations[id][cid] = e
		}
	}
	return result
}
