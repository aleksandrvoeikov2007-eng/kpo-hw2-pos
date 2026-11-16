package console

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"pos/internal/application"
	"pos/internal/domain"
	"strconv"
	"strings"
)

const Help = `Команды: поля разделяются символом ;
student;add;ID;имя
student;rename;ID;новое имя
student;remove;ID
students
criterion;add;ID;название;максимум;вес
criterion;update;ID;название;максимум;вес
criterion;remove;ID
criteria
grade;ID студента;ID критерия;балл   - выставить или исправить оценку
grades;ID студента
cancel;ID студента;ID критерия
total;ID студента
totals                                - все итоги и статистика
policy;zero|strict                    - правило для отсутствующих оценок
export;путь.csv                       - экспорт всех итогов
help
exit
Вес: 0 < вес <= 1. Дробная часть отделяется точкой. Итог: шкала 0–10.
По умолчанию отсутствующие оценки считаются нулевыми (zero).
Пустые строки и строки с # в начале пропускаются.`

type App struct{ facade *application.Facade }

func New(facade *application.Facade) *App { return &App{facade} }
func (a *App) Run(in io.Reader, out io.Writer) error {
	fmt.Fprintln(out, "ПОС — Преподавательская оценка студентов\n"+Help)
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "> ")
		if !scanner.Scan() {
			return scanner.Err()
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		reader := csv.NewReader(strings.NewReader(line))
		reader.Comma = ';'
		reader.TrimLeadingSpace = true
		fields, err := reader.Read()
		if err != nil {
			fmt.Fprintln(out, "Ошибка: неверный формат команды")
			continue
		}
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		if len(fields) == 1 && fields[0] == "exit" {
			return nil
		}
		if err := a.execute(fields, out); err != nil {
			fmt.Fprintln(out, "Ошибка:", err)
		}
	}
}
func count(fields []string, expected int) error {
	if len(fields) != expected {
		return fmt.Errorf("неверное число полей, введите help")
	}
	return nil
}
func number(text string) (float64, error) {
	n, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("%q: ожидается число с точкой", text)
	}
	return n, nil
}
func success(out io.Writer, err error) error {
	if err == nil {
		fmt.Fprintln(out, "Готово")
	}
	return err
}
func (a *App) execute(p []string, out io.Writer) error {
	switch p[0] {
	case "student":
		return a.student(p, out)
	case "criterion":
		return a.criterion(p, out)
	case "grade":
		if err := count(p, 4); err != nil {
			return err
		}
		score, err := number(p[3])
		if err != nil {
			return err
		}
		return success(out, a.facade.Grade(p[1], p[2], score))
	case "cancel":
		if err := count(p, 3); err != nil {
			return err
		}
		return success(out, a.facade.Cancel(p[1], p[2]))
	case "grades", "total", "policy", "export":
		if err := count(p, 2); err != nil {
			return err
		}
		switch p[0] {
		case "grades":
			list, err := a.facade.Evaluations(p[1])
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Fprintln(out, "Оценок нет")
			}
			for _, e := range list {
				fmt.Fprintf(out, "%s | %g\n", e.CriterionID(), e.Score().Value())
			}
		case "total":
			r, err := a.facade.Total(p[1])
			if err != nil {
				return err
			}
			printResult(out, r)
		case "policy":
			return success(out, a.facade.UsePolicy(p[1]))
		case "export":
			if err := a.facade.Export(p[1]); err != nil {
				return err
			}
			fmt.Fprintln(out, "Итоги экспортированы:", p[1])
		}
	default:
		if err := count(p, 1); err != nil {
			return err
		}
		switch p[0] {
		case "students":
			list, err := a.facade.Students()
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Fprintln(out, "Студентов нет")
			}
			for _, s := range list {
				fmt.Fprintf(out, "%s | %s\n", s.ID(), s.Name())
			}
		case "criteria":
			list, err := a.facade.Criteria()
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Fprintln(out, "Критериев нет")
			}
			for _, c := range list {
				fmt.Fprintf(out, "%s | %s | максимум %g | вес %g\n", c.ID(), c.Name(), c.Maximum(), c.Weight())
			}
		case "totals":
			report, err := a.facade.Totals()
			if err != nil {
				return err
			}
			for _, r := range report.Results {
				printResult(out, r)
			}
			if len(report.Results) == 0 {
				fmt.Fprintln(out, "Студентов нет")
				return nil
			}
			fmt.Fprintf(out, "Студентов: %d | средний %.2f | минимум %.2f | максимум %.2f\n", len(report.Results), report.Average, report.Minimum, report.Maximum)
		case "help":
			fmt.Fprintln(out, Help)
		default:
			return fmt.Errorf("неизвестная команда, введите help")
		}
	}
	return nil
}
func printResult(out io.Writer, r domain.Result) {
	fmt.Fprintf(out, "%s | %s | итог %.2f / 10 | оценок %d/%d\n", r.StudentID, r.StudentName, r.Score, r.Graded, r.Total)
}
func (a *App) student(p []string, out io.Writer) error {
	if len(p) < 2 {
		return fmt.Errorf("укажите действие student (help)")
	}
	switch p[1] {
	case "add", "rename":
		if err := count(p, 4); err != nil {
			return err
		}
		if p[1] == "add" {
			return success(out, a.facade.AddStudent(p[2], p[3]))
		}
		return success(out, a.facade.RenameStudent(p[2], p[3]))
	case "remove":
		if err := count(p, 3); err != nil {
			return err
		}
		return success(out, a.facade.RemoveStudent(p[2]))
	default:
		return fmt.Errorf("неизвестное действие student")
	}
}
func (a *App) criterion(p []string, out io.Writer) error {
	if len(p) < 2 {
		return fmt.Errorf("укажите действие criterion (help)")
	}
	switch p[1] {
	case "add", "update":
		if err := count(p, 6); err != nil {
			return err
		}
		maximum, err := number(p[4])
		if err != nil {
			return err
		}
		weight, err := number(p[5])
		if err != nil {
			return err
		}
		if p[1] == "add" {
			return success(out, a.facade.AddCriterion(p[2], p[3], maximum, weight))
		}
		return success(out, a.facade.UpdateCriterion(p[2], p[3], maximum, weight))
	case "remove":
		if err := count(p, 3); err != nil {
			return err
		}
		return success(out, a.facade.RemoveCriterion(p[2]))
	default:
		return fmt.Errorf("неизвестное действие criterion")
	}
}
