package csvexport

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"pos/internal/domain"
	"strconv"
	"strings"
)

type Exporter struct{}

func (Exporter) Export(path string, results []domain.Result) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("укажите путь CSV-файла")
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".pos-*.csv")
	if err != nil {
		return fmt.Errorf("создание экспорта: %w", err)
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	writer := csv.NewWriter(file)
	writer.Comma = ';'
	rows := [][]string{{"student_id", "student_name", "score", "graded", "total"}}
	for _, r := range results {
		rows = append(rows, []string{r.StudentID, r.StudentName, strconv.FormatFloat(r.Score, 'f', 2, 64), strconv.Itoa(r.Graded), strconv.Itoa(r.Total)})
	}
	writeErr := writer.WriteAll(rows)
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("запись экспорта: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("закрытие экспорта: %w", closeErr)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("сохранение экспорта: %w", err)
	}
	return nil
}
