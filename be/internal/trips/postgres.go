package trips

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("trip or itinerary item not found")
var ErrOrderChanged = errors.New("itinerary changed while reordering")

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) CreateTrip(ctx context.Context, userID string, t Trip) (Trip, error) {
	return scanTrip(s.pool.QueryRow(ctx, `INSERT INTO trips(user_id,title,start_date,end_date,origin,notes,status) VALUES($1,$2,$3::date,$4::date,$5,$6,$7) RETURNING id,title,start_date,end_date,origin,notes,status,created_at,updated_at`, userID, t.Title, t.StartDate, t.EndDate, t.Origin, t.Notes, t.Status))
}
func (s *PostgresStore) ListTrips(ctx context.Context, userID string) ([]Trip, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,title,start_date,end_date,origin,notes,status,created_at,updated_at FROM trips WHERE user_id=$1 ORDER BY start_date DESC,created_at DESC,id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Trip, 0)
	for rows.Next() {
		t, err := scanTrip(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}
func (s *PostgresStore) GetTrip(ctx context.Context, userID, id string) (Trip, error) {
	t, err := scanTrip(s.pool.QueryRow(ctx, `SELECT id,title,start_date,end_date,origin,notes,status,created_at,updated_at FROM trips WHERE user_id=$1 AND id=$2`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Trip{}, ErrNotFound
	}
	return t, err
}
func (s *PostgresStore) UpdateTrip(ctx context.Context, userID, id string, p TripPatch) (Trip, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback(ctx)
	current, err := scanTrip(tx.QueryRow(ctx, `SELECT id,title,start_date,end_date,origin,notes,status,created_at,updated_at FROM trips WHERE user_id=$1 AND id=$2 FOR UPDATE`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Trip{}, ErrNotFound
	}
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
	var outOfRange int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM itinerary_items WHERE trip_id=$1 AND scheduled_date IS NOT NULL AND (scheduled_date<$2::date OR scheduled_date>$3::date)`, id, current.StartDate, current.EndDate).Scan(&outOfRange); err != nil {
		return Trip{}, err
	}
	if outOfRange > 0 {
		return Trip{}, ErrValidation{"tanggal trip akan berada di luar jadwal item itinerary; perbarui jadwal item terlebih dahulu."}
	}
	updated, err := scanTrip(tx.QueryRow(ctx, `UPDATE trips SET title=$3,start_date=$4::date,end_date=$5::date,origin=$6,notes=$7,status=$8,updated_at=now() WHERE user_id=$1 AND id=$2 RETURNING id,title,start_date,end_date,origin,notes,status,created_at,updated_at`, userID, id, current.Title, current.StartDate, current.EndDate, current.Origin, current.Notes, current.Status))
	if err != nil {
		return Trip{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Trip{}, err
	}
	return updated, nil
}
func (s *PostgresStore) DeleteTrip(ctx context.Context, userID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM trips WHERE user_id=$1 AND id=$2`, userID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) CreateItem(ctx context.Context, userID, tripID string, item Item) (Item, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Item{}, err
	}
	defer tx.Rollback(ctx)
	var start, end pgtype.Date
	err = tx.QueryRow(ctx, `SELECT start_date,end_date FROM trips WHERE user_id=$1 AND id=$2 FOR UPDATE`, userID, tripID).Scan(&start, &end)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	if err != nil {
		return Item{}, err
	}
	if err := itemWithinTrip(item, start.Time, end.Time); err != nil {
		return Item{}, err
	}
	var sequence int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(sequence),0)+1 FROM itinerary_items WHERE trip_id=$1`, tripID).Scan(&sequence); err != nil {
		return Item{}, err
	}
	item, err = scanItem(tx.QueryRow(ctx, `INSERT INTO itinerary_items(trip_id,title,description,place_type,external_place_id,scheduled_date,start_time,end_time,sequence,transport_mode,notes) VALUES($1,$2,$3,$4,$5,$6::date,$7::time,$8::time,$9,$10,$11) RETURNING id,trip_id,title,description,place_type,external_place_id,scheduled_date,start_time,end_time,sequence,transport_mode,notes,created_at,updated_at`, tripID, item.Title, item.Description, item.PlaceType, item.ExternalPlaceID, item.ScheduledDate, item.StartTime, item.EndTime, sequence, item.TransportMode, item.Notes))
	if err != nil {
		return Item{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Item{}, err
	}
	return item, nil
}
func (s *PostgresStore) ListItems(ctx context.Context, userID, tripID string) ([]Item, error) {
	rows, err := s.pool.Query(ctx, `SELECT i.id,i.trip_id,i.title,i.description,i.place_type,i.external_place_id,i.scheduled_date,i.start_time,i.end_time,i.sequence,i.transport_mode,i.notes,i.created_at,i.updated_at FROM itinerary_items i JOIN trips t ON t.id=i.trip_id WHERE t.user_id=$1 AND t.id=$2 ORDER BY i.sequence,i.id`, userID, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		var exists bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM trips WHERE user_id=$1 AND id=$2)`, userID, tripID).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrNotFound
		}
	}
	return items, nil
}
func (s *PostgresStore) UpdateItem(ctx context.Context, userID, tripID, itemID string, p ItemPatch) (Item, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Item{}, err
	}
	defer tx.Rollback(ctx)
	var tripStart, tripEnd pgtype.Date
	err = tx.QueryRow(ctx, `SELECT start_date,end_date FROM trips WHERE user_id=$1 AND id=$2 FOR UPDATE`, userID, tripID).Scan(&tripStart, &tripEnd)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	if err != nil {
		return Item{}, err
	}
	current, err := scanItem(tx.QueryRow(ctx, `SELECT id,trip_id,title,description,place_type,external_place_id,scheduled_date,start_time,end_time,sequence,transport_mode,notes,created_at,updated_at FROM itinerary_items WHERE trip_id=$1 AND id=$2 FOR UPDATE`, tripID, itemID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	if err != nil {
		return Item{}, err
	}
	current = mergeItem(current, p)
	current, err = normalizeItem(current)
	if err != nil {
		return Item{}, err
	}
	if err := itemWithinTrip(current, tripStart.Time, tripEnd.Time); err != nil {
		return Item{}, err
	}
	updated, err := scanItem(tx.QueryRow(ctx, `UPDATE itinerary_items SET title=$3,description=$4,place_type=$5,external_place_id=$6,scheduled_date=$7::date,start_time=$8::time,end_time=$9::time,transport_mode=$10,notes=$11,updated_at=now() WHERE trip_id=$1 AND id=$2 RETURNING id,trip_id,title,description,place_type,external_place_id,scheduled_date,start_time,end_time,sequence,transport_mode,notes,created_at,updated_at`, tripID, itemID, current.Title, current.Description, current.PlaceType, current.ExternalPlaceID, current.ScheduledDate, current.StartTime, current.EndTime, current.TransportMode, current.Notes))
	if err != nil {
		return Item{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Item{}, err
	}
	return updated, nil
}
func (s *PostgresStore) DeleteItem(ctx context.Context, userID, tripID, itemID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var owned string
	err = tx.QueryRow(ctx, `SELECT id FROM trips WHERE user_id=$1 AND id=$2 FOR UPDATE`, userID, tripID).Scan(&owned)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM itinerary_items WHERE trip_id=$1 AND id=$2`, tripID, itemID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	rows, err := tx.Query(ctx, `SELECT id FROM itinerary_items WHERE trip_id=$1 ORDER BY sequence,id FOR UPDATE`, tripID)
	if err != nil {
		return err
	}
	remaining := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		remaining = append(remaining, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for n, id := range remaining {
		if _, err := tx.Exec(ctx, `UPDATE itinerary_items SET sequence=$3,updated_at=now() WHERE trip_id=$1 AND id=$2`, tripID, id, n+1); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (s *PostgresStore) ReorderItems(ctx context.Context, userID, tripID string, ids []string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var found string
	err = tx.QueryRow(ctx, `SELECT id FROM trips WHERE user_id=$1 AND id=$2 FOR UPDATE`, userID, tripID).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT id FROM itinerary_items WHERE trip_id=$1 ORDER BY id FOR UPDATE`, tripID)
	if err != nil {
		return err
	}
	existing := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		existing = append(existing, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if !sameIDs(existing, ids) {
		return ErrOrderChanged
	}
	for i, id := range ids {
		if _, err := tx.Exec(ctx, `UPDATE itinerary_items SET sequence=$3,updated_at=now() WHERE trip_id=$1 AND id=$2`, tripID, id, i+1); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type scanner interface{ Scan(...any) error }

func scanTrip(row scanner) (Trip, error) {
	var t Trip
	var start, end pgtype.Date
	err := row.Scan(&t.ID, &t.Title, &start, &end, &t.Origin, &t.Notes, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return Trip{}, err
	}
	t.StartDate = start.Time.Format("2006-01-02")
	t.EndDate = end.Time.Format("2006-01-02")
	return t, nil
}
func scanItem(row scanner) (Item, error) {
	var i Item
	var external pgtype.Text
	var date pgtype.Date
	var start, end pgtype.Time
	err := row.Scan(&i.ID, &i.TripID, &i.Title, &i.Description, &i.PlaceType, &external, &date, &start, &end, &i.Sequence, &i.TransportMode, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return Item{}, err
	}
	if external.Valid {
		v := external.String
		i.ExternalPlaceID = &v
	}
	if date.Valid {
		v := date.Time.Format("2006-01-02")
		i.ScheduledDate = &v
	}
	if start.Valid {
		v := formatDBTime(start.Microseconds)
		i.StartTime = &v
	}
	if end.Valid {
		v := formatDBTime(end.Microseconds)
		i.EndTime = &v
	}
	return i, nil
}
func formatDBTime(micros int64) string {
	seconds := micros / 1_000_000
	return fmt.Sprintf("%02d:%02d:%02d", seconds/3600, (seconds%3600)/60, seconds%60)
}
func itemWithinTrip(i Item, start, end time.Time) error {
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
func mergeItem(current Item, p ItemPatch) Item {
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
func sameIDs(existing, requested []string) bool {
	if len(existing) != len(requested) {
		return false
	}
	set := make(map[string]bool, len(existing))
	for _, id := range existing {
		set[id] = true
	}
	for _, id := range requested {
		if !set[id] {
			return false
		}
		delete(set, id)
	}
	return len(set) == 0
}
