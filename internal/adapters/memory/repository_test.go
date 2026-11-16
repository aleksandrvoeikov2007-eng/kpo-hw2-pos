package memory_test

import (
	"pos/internal/adapters/memory"
	"pos/internal/domain"
	"testing"
)

func TestRepositoryStoresAndReturnsCopies(t *testing.T) {
	r := memory.New()
	b := domain.NewGradebook()
	if err := b.AddStudent("S", "Анна"); err != nil {
		t.Fatal(err)
	}
	if err := r.Save(b); err != nil {
		t.Fatal(err)
	}
	if err := b.RenameStudent("S", "После сохранения"); err != nil {
		t.Fatal(err)
	}
	loaded, err := r.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Students()[0].Name() != "Анна" {
		t.Fatal("сохранение оставило внешнюю ссылку")
	}
	if err := loaded.RemoveStudent("S"); err != nil {
		t.Fatal(err)
	}
	again, err := r.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Students()) != 1 {
		t.Fatal("загрузка раскрывает хранимый агрегат")
	}
}
