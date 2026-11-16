package memory

import "pos/internal/domain"

type Repository struct{ book *domain.Gradebook }

func New() *Repository                                 { return &Repository{domain.NewGradebook()} }
func (r *Repository) Load() (*domain.Gradebook, error) { return r.book.Clone(), nil }
func (r *Repository) Save(book *domain.Gradebook) error {
	r.book = book.Clone()
	return nil
}
