package directory

import (
	"context"
	"errors"
)

type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return e.Message }

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }
func (s *Service) List(ctx context.Context, k Kind, f Filters) ([]Place, int, error) {
	return s.store.List(ctx, k, f)
}
func (s *Service) Get(ctx context.Context, k Kind, id string) (Place, error) {
	return s.store.Get(ctx, k, id)
}
func (s *Service) Create(ctx context.Context, k Kind, p Place) (Place, error) {
	normalized, err := NormalizePlace(k, p)
	if err != nil {
		return Place{}, ValidationError{Message: err.Error()}
	}
	return s.store.Create(ctx, k, normalized)
}
func (s *Service) Update(ctx context.Context, k Kind, id string, p Place) (Place, error) {
	normalized, err := NormalizePlace(k, p)
	if err != nil {
		return Place{}, ValidationError{Message: err.Error()}
	}
	return s.store.Update(ctx, k, id, normalized)
}
func (s *Service) Delete(ctx context.Context, k Kind, id string) error {
	return s.store.Delete(ctx, k, id)
}
func (s *Service) Save(ctx context.Context, userID string, k Kind, id string) error {
	return s.store.Save(ctx, userID, k, id)
}
func (s *Service) Saved(ctx context.Context, userID string) ([]SavedRef, error) {
	return s.store.Saved(ctx, userID)
}
func (s *Service) Unsave(ctx context.Context, userID string, k Kind, id string) error {
	return s.store.Unsave(ctx, userID, k, id)
}

func invalidInput(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}
