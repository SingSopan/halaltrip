package trips

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

func normalizeTrip(t Trip) (Trip, error) {
	t.Title = strings.TrimSpace(t.Title)
	t.Origin = strings.TrimSpace(t.Origin)
	if t.Title == "" || utf8.RuneCountInString(t.Title) > 160 {
		return t, ErrValidation{"title wajib diisi dan maksimal 160 karakter."}
	}
	start, err := time.Parse("2006-01-02", t.StartDate)
	if err != nil || start.Format("2006-01-02") != t.StartDate {
		return t, ErrValidation{"start_date harus berupa tanggal YYYY-MM-DD."}
	}
	end, err := time.Parse("2006-01-02", t.EndDate)
	if err != nil || end.Format("2006-01-02") != t.EndDate {
		return t, ErrValidation{"end_date harus berupa tanggal YYYY-MM-DD."}
	}
	if end.Before(start) {
		return t, ErrValidation{"end_date tidak boleh lebih awal dari start_date."}
	}
	if len(t.Origin) > 300 || len(t.Notes) > 5000 {
		return t, ErrValidation{"origin atau notes terlalu panjang."}
	}
	if t.Status == "" {
		t.Status = StatusPlanning
	}
	switch t.Status {
	case StatusPlanning, StatusActive, StatusCompleted, StatusCancelled:
	default:
		return t, ErrValidation{"status trip tidak valid."}
	}
	return t, nil
}

func normalizeItem(i Item) (Item, error) {
	i.Title = strings.TrimSpace(i.Title)
	if i.Title == "" || utf8.RuneCountInString(i.Title) > 200 {
		return i, ErrValidation{"title item wajib diisi dan maksimal 200 karakter."}
	}
	switch i.PlaceType {
	case "destination", "restaurant", "mosque", "hotel", "custom_activity":
	default:
		return i, ErrValidation{"place_type itinerary tidak valid."}
	}
	if i.ExternalPlaceID != nil {
		v := strings.TrimSpace(*i.ExternalPlaceID)
		if len(v) > 256 {
			return i, ErrValidation{"external_place_id terlalu panjang."}
		}
		if v == "" {
			i.ExternalPlaceID = nil
		} else {
			i.ExternalPlaceID = &v
		}
	}
	if i.ScheduledDate != nil {
		v := *i.ScheduledDate
		if v == "" {
			i.ScheduledDate = nil
		} else if d, err := time.Parse("2006-01-02", v); err != nil || d.Format("2006-01-02") != v {
			return i, ErrValidation{"scheduled_date harus berupa tanggal YYYY-MM-DD."}
		}
	}
	if i.StartTime != nil && *i.StartTime == "" {
		i.StartTime = nil
	}
	if i.EndTime != nil && *i.EndTime == "" {
		i.EndTime = nil
	}
	if (i.StartTime == nil) != (i.EndTime == nil) {
		return i, ErrValidation{"start_time dan end_time harus diisi bersamaan."}
	}
	if i.StartTime != nil {
		start, err := parseClock(*i.StartTime)
		if err != nil {
			return i, ErrValidation{"start_time harus berupa waktu HH:MM atau HH:MM:SS."}
		}
		end, err := parseClock(*i.EndTime)
		if err != nil {
			return i, ErrValidation{"end_time harus berupa waktu HH:MM atau HH:MM:SS."}
		}
		if !end.After(start) {
			return i, ErrValidation{"end_time harus lebih akhir dari start_time."}
		}
		a, b := start.Format("15:04:05"), end.Format("15:04:05")
		i.StartTime = &a
		i.EndTime = &b
	}
	if utf8.RuneCountInString(i.Description) > 5000 || utf8.RuneCountInString(i.TransportMode) > 80 || utf8.RuneCountInString(i.Notes) > 3000 {
		return i, ErrValidation{"field itinerary terlalu panjang."}
	}
	return i, nil
}
func parseClock(v string) (time.Time, error) {
	if t, err := time.Parse("15:04", v); err == nil {
		return t, nil
	}
	return time.Parse("15:04:05", v)
}

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }
func (s *Service) CreateTrip(ctx context.Context, user string, t Trip) (Trip, error) {
	t, err := normalizeTrip(t)
	if err != nil {
		return Trip{}, err
	}
	return s.store.CreateTrip(ctx, user, t)
}
func (s *Service) ListTrips(ctx context.Context, user string) ([]Trip, error) {
	return s.store.ListTrips(ctx, user)
}
func (s *Service) GetTrip(ctx context.Context, user, id string) (Trip, error) {
	return s.store.GetTrip(ctx, user, id)
}
func (s *Service) UpdateTrip(ctx context.Context, user, id string, p TripPatch) (Trip, error) {
	current, err := s.store.GetTrip(ctx, user, id)
	if err != nil {
		return Trip{}, err
	}
	if p.Title != nil {
		current.Title = *p.Title
	}
	if p.StartDate != nil {
		current.StartDate = *p.StartDate
	}
	if p.EndDate != nil {
		current.EndDate = *p.EndDate
	}
	if p.Origin != nil {
		current.Origin = *p.Origin
	}
	if p.Notes != nil {
		current.Notes = *p.Notes
	}
	if p.Status != nil {
		current.Status = *p.Status
	}
	current, err = normalizeTrip(current)
	if err != nil {
		return Trip{}, err
	}
	items, err := s.store.ListItems(ctx, user, id)
	if err != nil {
		return Trip{}, err
	}
	start, _ := time.Parse("2006-01-02", current.StartDate)
	end, _ := time.Parse("2006-01-02", current.EndDate)
	for _, item := range items {
		if err := itemWithin(item, start, end); err != nil {
			return Trip{}, err
		}
	}
	return s.store.UpdateTrip(ctx, user, id, p)
}
func (s *Service) DeleteTrip(ctx context.Context, user, id string) error {
	return s.store.DeleteTrip(ctx, user, id)
}
func (s *Service) CreateItem(ctx context.Context, user, tripID string, item Item) (Item, error) {
	trip, err := s.store.GetTrip(ctx, user, tripID)
	if err != nil {
		return Item{}, err
	}
	item, err = normalizeItem(item)
	if err != nil {
		return Item{}, err
	}
	start, _ := time.Parse("2006-01-02", trip.StartDate)
	end, _ := time.Parse("2006-01-02", trip.EndDate)
	if err := itemWithin(item, start, end); err != nil {
		return Item{}, err
	}
	return s.store.CreateItem(ctx, user, tripID, item)
}
func (s *Service) ListItems(ctx context.Context, user, tripID string) ([]Item, error) {
	return s.store.ListItems(ctx, user, tripID)
}
func (s *Service) UpdateItem(ctx context.Context, user, tripID, itemID string, p ItemPatch) (Item, error) {
	trip, err := s.store.GetTrip(ctx, user, tripID)
	if err != nil {
		return Item{}, err
	}
	items, err := s.store.ListItems(ctx, user, tripID)
	if err != nil {
		return Item{}, err
	}
	var current *Item
	for i := range items {
		if items[i].ID == itemID {
			current = &items[i]
			break
		}
	}
	if current == nil {
		return Item{}, ErrNotFound
	}
	merged := merge(*current, p)
	merged, err = normalizeItem(merged)
	if err != nil {
		return Item{}, err
	}
	start, _ := time.Parse("2006-01-02", trip.StartDate)
	end, _ := time.Parse("2006-01-02", trip.EndDate)
	if err := itemWithin(merged, start, end); err != nil {
		return Item{}, err
	}
	return s.store.UpdateItem(ctx, user, tripID, itemID, p)
}
func (s *Service) DeleteItem(ctx context.Context, user, tripID, itemID string) error {
	return s.store.DeleteItem(ctx, user, tripID, itemID)
}
func (s *Service) ReorderItems(ctx context.Context, user, tripID string, ids []string) error {
	items, err := s.store.ListItems(ctx, user, tripID)
	if err != nil {
		return err
	}
	if len(ids) != len(items) {
		return ErrValidation{"order harus memuat setiap item itinerary tepat satu kali."}
	}
	valid := make(map[string]bool, len(items))
	for _, item := range items {
		valid[item.ID] = true
	}
	for _, id := range ids {
		if !valid[id] {
			return ErrValidation{"order memuat item yang tidak ada atau duplikat."}
		}
		delete(valid, id)
	}
	if len(valid) != 0 {
		return ErrValidation{"order harus memuat setiap item itinerary."}
	}
	return s.store.ReorderItems(ctx, user, tripID, ids)
}

func merge(current Item, p ItemPatch) Item {
	if p.Title != nil {
		current.Title = *p.Title
	}
	if p.Description != nil {
		current.Description = *p.Description
	}
	if p.PlaceType != nil {
		current.PlaceType = *p.PlaceType
	}
	if p.ExternalPlaceID != nil {
		v := *p.ExternalPlaceID
		if v == "" {
			current.ExternalPlaceID = nil
		} else {
			current.ExternalPlaceID = &v
		}
	}
	if p.ScheduledDate != nil {
		v := *p.ScheduledDate
		if v == "" {
			current.ScheduledDate = nil
		} else {
			current.ScheduledDate = &v
		}
	}
	if p.StartTime != nil {
		v := *p.StartTime
		if v == "" {
			current.StartTime = nil
		} else {
			current.StartTime = &v
		}
	}
	if p.EndTime != nil {
		v := *p.EndTime
		if v == "" {
			current.EndTime = nil
		} else {
			current.EndTime = &v
		}
	}
	if p.TransportMode != nil {
		current.TransportMode = *p.TransportMode
	}
	if p.Notes != nil {
		current.Notes = *p.Notes
	}
	return current
}
func itemWithin(i Item, start, end time.Time) error {
	if i.ScheduledDate == nil {
		return nil
	}
	d, err := time.Parse("2006-01-02", *i.ScheduledDate)
	if err != nil {
		return ErrValidation{"scheduled_date harus berupa tanggal YYYY-MM-DD."}
	}
	if d.Before(start) || d.After(end) {
		return ErrValidation{"scheduled_date harus berada dalam rentang perjalanan."}
	}
	return nil
}
func IsValidation(err error) bool { var target ErrValidation; return errors.As(err, &target) }
