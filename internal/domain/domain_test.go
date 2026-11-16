package domain_test

import (
	"math"
	"pos/internal/domain"
	"testing"
)

func requireOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func requireError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("ожидалась ошибка")
	}
}
func near(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("получено %g, ожидалось %g", got, want)
	}
}
func book(t *testing.T) *domain.Gradebook {
	t.Helper()
	b := domain.NewGradebook()
	requireOK(t, b.AddStudent("S1", "Анна"))
	requireOK(t, b.AddCriterion("C1", "Практика", 10, 0.4))
	requireOK(t, b.AddCriterion("C2", "Экзамен", 20, 0.6))
	return b
}

func TestValueObjectsAndEntities(t *testing.T) {
	s, err := domain.NewStudent(" S1 ", " Анна ")
	requireOK(t, err)
	if s.ID() != "S1" || s.Name() != "Анна" {
		t.Fatal(s)
	}
	for _, p := range [][2]string{{"", "Анна"}, {"S1", " "}} {
		_, err := domain.NewStudent(p[0], p[1])
		requireError(t, err)
	}
	for _, maximum := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		_, err := domain.NewCriterion("C", "Имя", maximum, 1)
		requireError(t, err)
	}
	for _, weight := range []float64{0, -1, 1.01, math.NaN(), math.Inf(1)} {
		_, err := domain.NewCriterion("C", "Имя", 10, weight)
		requireError(t, err)
	}
	_, err = domain.NewCriterion("", "Имя", 10, 1)
	requireError(t, err)
	c, err := domain.NewCriterion(" C ", " Имя ", 10, 1)
	requireOK(t, err)
	if c.ID() != "C" || c.Name() != "Имя" || c.Maximum() != 10 || c.Weight() != 1 {
		t.Fatal(c)
	}
	for _, value := range []float64{0, 5.5, 10} {
		score, err := domain.NewScore(value, 10)
		requireOK(t, err)
		near(t, score.Value(), value)
	}
	for _, value := range []float64{-1, 10.01, math.NaN(), math.Inf(1)} {
		_, err := domain.NewScore(value, 10)
		requireError(t, err)
	}
	for _, maximum := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		_, err := domain.NewScore(0, maximum)
		requireError(t, err)
	}
}
func TestFactoryChecksRubricAndScore(t *testing.T) {
	b := book(t)
	factory := domain.EvaluationFactory{}
	for _, value := range []float64{-1, 11, math.NaN(), math.Inf(-1)} {
		_, err := factory.Create(b.Criteria(), "C1", value)
		requireError(t, err)
	}
	_, err := factory.Create(b.Criteria(), "missing", 0)
	requireError(t, err)
	evaluation, err := factory.Create(b.Criteria(), "C1", 10)
	requireOK(t, err)
	if evaluation.CriterionID() != "C1" || evaluation.Score().Value() != 10 {
		t.Fatal(evaluation)
	}
}
func TestGradeCorrectionCancellationAndCurrentTotal(t *testing.T) {
	b := book(t)
	calculator := domain.NewCalculator(domain.MissingAsZero{})
	requireOK(t, b.Grade("S1", "C1", 8))
	requireOK(t, b.Grade("S1", "C2", 15))
	r, err := calculator.Calculate(b, "S1")
	requireOK(t, err)
	near(t, r.Score, 7.7)
	if r.StudentID != "S1" || r.StudentName != "Анна" || r.Graded != 2 || r.Total != 2 {
		t.Fatal(r)
	}
	requireError(t, b.Grade("S1", "C1", 12))
	requireError(t, b.Grade("S1", "missing", 1))
	requireError(t, b.Grade("missing", "C1", 1))
	r, err = calculator.Calculate(b, "S1")
	requireOK(t, err)
	near(t, r.Score, 7.7)
	requireOK(t, b.Grade("S1", "C1", 10))
	r, err = calculator.Calculate(b, "S1")
	requireOK(t, err)
	near(t, r.Score, 8.5)
	requireOK(t, b.Cancel("S1", "C2"))
	r, err = calculator.Calculate(b, "S1")
	requireOK(t, err)
	near(t, r.Score, 4)
	requireError(t, b.Cancel("S1", "C2"))
	requireError(t, b.Cancel("missing", "C1"))
	requireOK(t, b.Grade("S1", "C1", 0))
	r, err = calculator.Calculate(b, "S1")
	requireOK(t, err)
	near(t, r.Score, 0)
}
func TestStrategiesAndNormalizedWeights(t *testing.T) {
	b := book(t)
	strict := domain.NewCalculator(domain.RequireComplete{})
	_, err := strict.Calculate(b, "S1")
	requireError(t, err)
	requireOK(t, b.Grade("S1", "C1", 8))
	_, err = strict.Calculate(b, "S1")
	requireError(t, err)
	requireOK(t, b.Grade("S1", "C2", 15))
	r, err := strict.Calculate(b, "S1")
	requireOK(t, err)
	near(t, r.Score, 7.7)
	requireOK(t, b.UpdateCriterion("C1", "Практика", 10, 0.2))
	requireOK(t, b.UpdateCriterion("C2", "Экзамен", 20, 0.3))
	r, err = strict.Calculate(b, "S1")
	requireOK(t, err)
	near(t, r.Score, 7.7)
	requireOK(t, b.UpdateCriterion("C1", "Практика", 20, 0.2))
	r, err = strict.Calculate(b, "S1")
	requireOK(t, err)
	near(t, r.Score, 6.1)
	_, err = strict.Calculate(b, "missing")
	requireError(t, err)
	empty := domain.NewGradebook()
	requireOK(t, empty.AddStudent("S", "Студент"))
	_, err = strict.Calculate(empty, "S")
	requireError(t, err)
}
func TestStudentAndCriterionLifecycle(t *testing.T) {
	b := book(t)
	requireError(t, b.AddStudent("S1", "Дубликат"))
	requireError(t, b.AddStudent("", "Имя"))
	requireError(t, b.RenameStudent("missing", "Имя"))
	requireError(t, b.RenameStudent("S1", ""))
	requireOK(t, b.RenameStudent("S1", "Новое имя"))
	s, err := b.Student("S1")
	requireOK(t, err)
	if s.Name() != "Новое имя" {
		t.Fatal(s)
	}
	requireError(t, b.AddCriterion("C1", "Повтор", 10, 1))
	requireError(t, b.AddCriterion("C3", "Неверный", 0, 1))
	requireError(t, b.UpdateCriterion("missing", "Имя", 10, 1))
	requireError(t, b.UpdateCriterion("C1", "Имя", 10, 0))
	requireOK(t, b.Grade("S1", "C1", 8))
	requireError(t, b.UpdateCriterion("C1", "Имя", 7, 0.4))
	if b.Criteria()[0].Maximum() != 10 {
		t.Fatal("неудачное изменение затронуло критерий")
	}
	requireError(t, b.RemoveCriterion("C1"))
	requireError(t, b.RemoveCriterion("missing"))
	requireOK(t, b.RemoveStudent("S1"))
	requireError(t, b.RemoveStudent("S1"))
	_, err = b.Evaluations("S1")
	requireError(t, err)
	requireOK(t, b.RemoveCriterion("C1"))
	if len(b.Students()) != 0 || len(b.Criteria()) != 1 {
		t.Fatal("удаление не выполнено")
	}
}
func TestCopiesDoNotExposeAggregateState(t *testing.T) {
	b := book(t)
	requireOK(t, b.Grade("S1", "C1", 8))
	clone := b.Clone()
	requireOK(t, clone.Grade("S1", "C1", 1))
	requireOK(t, clone.RenameStudent("S1", "Другое имя"))
	requireOK(t, clone.UpdateCriterion("C1", "Другое имя", 20, 1))
	evaluations, err := b.Evaluations("S1")
	requireOK(t, err)
	if evaluations[0].Score().Value() != 8 {
		t.Fatal("копия меняет исходную ведомость")
	}
	evaluations[0] = domain.Evaluation{}
	students := b.Students()
	students[0] = domain.Student{}
	criteria := b.Criteria()
	criteria[0] = domain.Criterion{}
	current, err := b.Evaluations("S1")
	requireOK(t, err)
	if current[0].Score().Value() != 8 || b.Students()[0].Name() != "Анна" || b.Criteria()[0].Maximum() != 10 {
		t.Fatal("списки раскрывают внутреннее состояние")
	}
}
