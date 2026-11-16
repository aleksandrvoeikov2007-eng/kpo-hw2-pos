package console_test

import (
	"bytes"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"pos/internal/adapters/console"
	"pos/internal/di"
	"strings"
	"testing"
)

func TestEndToEndDemo(t *testing.T) {
	commands, err := os.ReadFile("../../../examples/demo.txt")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "results.csv")
	input := strings.ReplaceAll(string(commands), "export;results.csv", "export;"+path)
	var out bytes.Buffer
	if err := console.New(di.Default().Facade()).Run(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Студентов нет", "Критериев нет", "S1 | Анна Иванова", "C1 | Практика | максимум 10 | вес 0.4", "C1 | 8",
		"итог 7.70 / 10", "итог 6.00 / 10", "S1 | Анна Иванова | итог 9.60 / 10 | оценок 2/2",
		"S2 | Борис Сергеев | итог 10.00 / 10 | оценок 2/2", "средний 9.80", "Итоги экспортированы:",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("нет %q в выводе:\n%s", want, out.String())
		}
	}
	if got := strings.Count(out.String(), "Ошибка:"); got != 10 {
		t.Fatalf("ошибок %d, ожидалось 10:\n%s", got, out.String())
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[1][0] != "S1" || rows[1][2] != "9.60" || rows[2][1] != "Борис Сергеев" || rows[2][2] != "10.00" {
		t.Fatal(rows)
	}
}
func TestInvalidCommandsAndRecovery(t *testing.T) {
	commands := []string{
		"unknown", "students;extra", "student", "student;unknown", "student;add;S", "student;rename;S", "student;remove",
		"criterion", "criterion;unknown", "criterion;add", "criterion;remove", "criterion;add;C;Имя;abc;1", "criterion;add;C;Имя;10;abc",
		"criterion;add;C;Имя;NaN;1", "criterion;add;C;Имя;10;Inf", "grade", "grade;S;C;abc", "grade;S;C;10",
		"cancel", "cancel;S;C", "total", "total;missing", "grades;missing", "policy;unknown", "export;", "export;missing-dir/results.csv",
		`student;add;S;"незакрытая кавычка`,
	}
	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			var out bytes.Buffer
			input := command + "\nstudent;add;SAFE;Успех\nstudents\nexit\n"
			if err := console.New(di.Default().Facade()).Run(strings.NewReader(input), &out); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), "Ошибка:") || !strings.Contains(out.String(), "SAFE | Успех") {
				t.Fatal(out.String())
			}
		})
	}
}
func TestEmptyViewsAndQuotedFields(t *testing.T) {
	var out bytes.Buffer
	input := "\n # комментарий\nstudent;add;S; \"Анна; Иванова\"\ncriterion;add;C;Критерий;10;1\ngrades;S\nhelp\nstudent;remove;S\ntotals\nexit\nstudent;add;AFTER;После выхода\n"
	c := di.Default()
	if err := console.New(c.Facade()).Run(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Оценок нет", "Студентов нет"} {
		if !strings.Contains(out.String(), want) {
			t.Fatal(out.String())
		}
	}
	list, err := c.Facade().Students()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 || strings.Contains(out.String(), "Ошибка:") {
		t.Fatal(out.String())
	}
}
func TestStrictExportDoesNotOverwriteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "results.csv")
	if err := os.WriteFile(path, []byte("старый результат"), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	input := "student;add;S;Анна\ncriterion;add;C;Критерий;10;1\npolicy;strict\nexport;" + path + "\n"
	if err := console.New(di.Default().Facade()).Run(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "старый результат" || !strings.Contains(out.String(), "Ошибка:") {
		t.Fatal(out.String())
	}
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("ошибка чтения") }
func TestReadErrors(t *testing.T) {
	for _, input := range []interface{ Read([]byte) (int, error) }{brokenReader{}, strings.NewReader(strings.Repeat("x", 70000))} {
		var out bytes.Buffer
		if err := console.New(di.Default().Facade()).Run(input, &out); err == nil {
			t.Fatal("ошибка чтения не передана вызывающему коду")
		}
	}
}
