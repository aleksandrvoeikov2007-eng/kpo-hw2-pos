package csvexport_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"pos/internal/adapters/csvexport"
	"pos/internal/domain"
	"strings"
	"testing"
)

func TestCSVExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "results.csv")
	exporter := csvexport.Exporter{}
	results := []domain.Result{{StudentID: "S1", StudentName: "Анна; \"Иванова\"", Score: 7.7, Graded: 2, Total: 3}}
	if err := exporter.Export(path, results); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1][1] != results[0].StudentName || rows[1][2] != "7.70" || rows[1][3] != "2" || rows[1][4] != "3" {
		t.Fatal(rows)
	}
	if err := exporter.Export(path, nil); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "\n") != 1 {
		t.Fatal("перезапись оставила старые строки")
	}
}
func TestExportErrorsAndTemporaryCleanup(t *testing.T) {
	dir := t.TempDir()
	exporter := csvexport.Exporter{}
	for _, path := range []string{"", filepath.Join(dir, "missing", "x.csv"), dir} {
		if err := exporter.Export(path, nil); err == nil {
			t.Fatal("неверный путь принят", path)
		}
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatal("временный файл не удалён", files)
	}
}
