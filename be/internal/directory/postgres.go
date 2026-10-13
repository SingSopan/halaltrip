package directory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("place not found")

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func validKind(k Kind) bool { return k == Restaurant || k == Mosque || k == Hotel }
func table(k Kind) string {
	switch k {
	case Restaurant:
		return "restaurants"
	case Mosque:
		return "mosques"
	case Hotel:
		return "hotels"
	default:
		return ""
	}
}

func (s *PostgresStore) List(ctx context.Context, k Kind, f Filters) ([]Place, int, error) {
	t := table(k)
	if t == "" {
		return nil, 0, fmt.Errorf("invalid place kind")
	}
	where, args := buildWhere(k, f)
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM "+t+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	columns := columns(k)
	orderCol := map[string]string{"name": "name", "city": "city", "created_at": "created_at", "updated_at": "updated_at"}[f.Sort]
	if orderCol == "" {
		orderCol = "name"
	}
	direction := "ASC"
	if f.Order == "desc" {
		direction = "DESC"
	}
	query := fmt.Sprintf("SELECT %s FROM %s%s ORDER BY %s %s, id ASC LIMIT $%d OFFSET $%d", columns, t, where, orderCol, direction, len(args)+1, len(args)+2)
	args = append(args, f.Limit, f.Offset)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	result := make([]Place, 0)
	for rows.Next() {
		p, err := scanPlace(rows, k)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, p)
	}
	return result, total, rows.Err()
}
func (s *PostgresStore) Get(ctx context.Context, k Kind, id string) (Place, error) {
	t := table(k)
	if t == "" {
		return Place{}, fmt.Errorf("invalid place kind")
	}
	p, err := scanPlace(s.pool.QueryRow(ctx, "SELECT "+columns(k)+" FROM "+t+" WHERE id=$1", id), k)
	if errors.Is(err, pgx.ErrNoRows) {
		return Place{}, ErrNotFound
	}
	return p, err
}
func (s *PostgresStore) Create(ctx context.Context, k Kind, p Place) (Place, error) {
	t := table(k)
	if t == "" {
		return Place{}, fmt.Errorf("invalid place kind")
	}
	var query string
	var args []any
	switch k {
	case Restaurant:
		query = `INSERT INTO restaurants(name,description,city,country,address,latitude,longitude,image_url,cuisine,price_range,halal_status,halal_verification_note,source_url) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING ` + columns(k)
		args = []any{p.Name, p.Description, p.City, p.Country, p.Address, p.Latitude, p.Longitude, p.ImageURL, p.Cuisine, p.PriceRange, p.HalalStatus, p.HalalVerificationNote, p.SourceURL}
	case Mosque:
		facilities, _ := json.Marshal(p.Facilities)
		query = `INSERT INTO mosques(name,description,city,country,address,latitude,longitude,image_url,facilities,source_url) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING ` + columns(k)
		args = []any{p.Name, p.Description, p.City, p.Country, p.Address, p.Latitude, p.Longitude, p.ImageURL, facilities, p.SourceURL}
	case Hotel:
		features, _ := json.Marshal(p.MuslimFriendlyFeatures)
		query = `INSERT INTO hotels(name,description,city,country,address,latitude,longitude,image_url,price_range,muslim_friendly_features,source_url) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING ` + columns(k)
		args = []any{p.Name, p.Description, p.City, p.Country, p.Address, p.Latitude, p.Longitude, p.ImageURL, p.PriceRange, features, p.SourceURL}
	}
	return scanPlace(s.pool.QueryRow(ctx, query, args...), k)
}
func (s *PostgresStore) Update(ctx context.Context, k Kind, id string, p Place) (Place, error) {
	t := table(k)
	if t == "" {
		return Place{}, fmt.Errorf("invalid place kind")
	}
	var query string
	var args []any
	switch k {
	case Restaurant:
		query = `UPDATE restaurants SET name=$2,description=$3,city=$4,country=$5,address=$6,latitude=$7,longitude=$8,image_url=$9,cuisine=$10,price_range=$11,halal_status=$12,halal_verification_note=$13,source_url=$14,updated_at=now() WHERE id=$1 RETURNING ` + columns(k)
		args = []any{id, p.Name, p.Description, p.City, p.Country, p.Address, p.Latitude, p.Longitude, p.ImageURL, p.Cuisine, p.PriceRange, p.HalalStatus, p.HalalVerificationNote, p.SourceURL}
	case Mosque:
		facilities, _ := json.Marshal(p.Facilities)
		query = `UPDATE mosques SET name=$2,description=$3,city=$4,country=$5,address=$6,latitude=$7,longitude=$8,image_url=$9,facilities=$10,source_url=$11,updated_at=now() WHERE id=$1 RETURNING ` + columns(k)
		args = []any{id, p.Name, p.Description, p.City, p.Country, p.Address, p.Latitude, p.Longitude, p.ImageURL, facilities, p.SourceURL}
	case Hotel:
		features, _ := json.Marshal(p.MuslimFriendlyFeatures)
		query = `UPDATE hotels SET name=$2,description=$3,city=$4,country=$5,address=$6,latitude=$7,longitude=$8,image_url=$9,price_range=$10,muslim_friendly_features=$11,source_url=$12,updated_at=now() WHERE id=$1 RETURNING ` + columns(k)
		args = []any{id, p.Name, p.Description, p.City, p.Country, p.Address, p.Latitude, p.Longitude, p.ImageURL, p.PriceRange, features, p.SourceURL}
	}
	place, err := scanPlace(s.pool.QueryRow(ctx, query, args...), k)
	if errors.Is(err, pgx.ErrNoRows) {
		return Place{}, ErrNotFound
	}
	return place, err
}
func (s *PostgresStore) Delete(ctx context.Context, k Kind, id string) error {
	t := table(k)
	if t == "" {
		return fmt.Errorf("invalid place kind")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM saved_places WHERE place_type=$1 AND place_id=$2`, string(k), id); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, "DELETE FROM "+t+" WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}
func (s *PostgresStore) Save(ctx context.Context, userID string, k Kind, id string) error {
	if _, err := s.Get(ctx, k, id); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO saved_places(user_id,place_type,place_id) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, userID, string(k), id)
	return err
}
func (s *PostgresStore) Saved(ctx context.Context, userID string) ([]SavedRef, error) {
	rows, err := s.pool.Query(ctx, `SELECT place_type,place_id,created_at FROM saved_places WHERE user_id=$1 ORDER BY created_at DESC,place_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refs := make([]SavedRef, 0)
	for rows.Next() {
		var ref SavedRef
		if err := rows.Scan(&ref.Kind, &ref.ID, &ref.CreatedAt); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}
func (s *PostgresStore) Unsave(ctx context.Context, userID string, k Kind, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM saved_places WHERE user_id=$1 AND place_type=$2 AND place_id=$3`, userID, string(k), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type rowScanner interface{ Scan(...any) error }

func columns(k Kind) string {
	base := `id,name,description,city,country,address,latitude,longitude,image_url,`
	switch k {
	case Restaurant:
		return base + `cuisine,price_range,halal_status,halal_verification_note,source_url,created_at,updated_at`
	case Mosque:
		return base + `facilities,source_url,created_at,updated_at`
	case Hotel:
		return base + `price_range,muslim_friendly_features,source_url,created_at,updated_at`
	default:
		return ""
	}
}
func scanPlace(row rowScanner, k Kind) (Place, error) {
	var p Place
	var lat, lon pgtype.Float8
	base := []any{&p.ID, &p.Name, &p.Description, &p.City, &p.Country, &p.Address, &lat, &lon, &p.ImageURL}
	var err error
	switch k {
	case Restaurant:
		err = row.Scan(append(base, &p.Cuisine, &p.PriceRange, &p.HalalStatus, &p.HalalVerificationNote, &p.SourceURL, &p.CreatedAt, &p.UpdatedAt)...)
	case Mosque:
		var raw []byte
		err = row.Scan(append(base, &raw, &p.SourceURL, &p.CreatedAt, &p.UpdatedAt)...)
		if err == nil {
			err = json.Unmarshal(raw, &p.Facilities)
		}
	case Hotel:
		var raw []byte
		err = row.Scan(append(base, &p.PriceRange, &raw, &p.SourceURL, &p.CreatedAt, &p.UpdatedAt)...)
		if err == nil {
			err = json.Unmarshal(raw, &p.MuslimFriendlyFeatures)
		}
	default:
		return Place{}, fmt.Errorf("invalid place kind")
	}
	if err != nil {
		return Place{}, err
	}
	if lat.Valid {
		p.Latitude = &lat.Float64
	}
	if lon.Valid {
		p.Longitude = &lon.Float64
	}
	return p, nil
}

func buildWhere(k Kind, f Filters) (string, []any) {
	clauses := []string{"TRUE"}
	args := make([]any, 0, 8)
	add := func(clause string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}
	if f.Query != "" {
		add("name ILIKE '%' || $%d || '%'", f.Query)
	}
	if f.City != "" {
		add("city ILIKE $%d", f.City)
	}
	if f.Country != "" {
		add("country ILIKE $%d", f.Country)
	}
	switch k {
	case Restaurant:
		if f.Cuisine != "" {
			add("cuisine ILIKE $%d", f.Cuisine)
		}
		if f.HalalStatus != "" {
			add("halal_status=$%d", f.HalalStatus)
		}
		if f.PriceRange != "" {
			add("price_range=$%d", f.PriceRange)
		}
	case Mosque:
		if f.Facility != "" {
			add("EXISTS (SELECT 1 FROM jsonb_array_elements_text(facilities) AS item(value) WHERE lower(item.value)=lower($%d))", f.Facility)
		}
	case Hotel:
		if f.PriceRange != "" {
			add("price_range=$%d", f.PriceRange)
		}
		if f.Feature != "" {
			add("EXISTS (SELECT 1 FROM jsonb_array_elements(muslim_friendly_features) AS item WHERE lower(item->>'name')=lower($%d))", f.Feature)
		}
	}
	return " WHERE " + joinAnd(clauses), args
}
func joinAnd(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += " AND "
		}
		out += p
	}
	return out
}
