package users

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("user not found")
var ErrConflict = errors.New("email already registered")

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type Preferences struct {
	Interests  []string `json:"interests"`
	Activity   string   `json:"activity"`
	TravelType string   `json:"travel_type"`
}
type Store interface {
	Create(context.Context, string, string, string) (User, error)
	ByEmail(context.Context, string) (User, string, error)
	ByID(context.Context, string) (User, error)
	Update(context.Context, string, *string, *string) (User, error)
	GetPreferences(context.Context, string) (Preferences, error)
	SavePreferences(context.Context, string, Preferences) error
}
type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) Create(ctx context.Context, name, email, hash string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `INSERT INTO users (name,email,password_hash) VALUES ($1,$2,$3) RETURNING id,name,email,created_at,updated_at`, name, email, hash).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, ErrConflict
	}
	return u, err
}
func (s *PostgresStore) ByEmail(ctx context.Context, email string) (User, string, error) {
	var u User
	var hash string
	err := s.pool.QueryRow(ctx, `SELECT id,name,email,password_hash,created_at,updated_at FROM users WHERE email=$1`, strings.ToLower(email)).Scan(&u.ID, &u.Name, &u.Email, &hash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, "", ErrNotFound
	}
	return u, hash, err
}
func (s *PostgresStore) ByID(ctx context.Context, id string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `SELECT id,name,email,created_at,updated_at FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}
func (s *PostgresStore) Update(ctx context.Context, id string, name, email *string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `UPDATE users SET name=COALESCE($2,name), email=COALESCE($3,email), updated_at=now() WHERE id=$1 RETURNING id,name,email,created_at,updated_at`, id, name, email).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, ErrConflict
	}
	return u, err
}
func (s *PostgresStore) GetPreferences(ctx context.Context, id string) (Preferences, error) {
	var p Preferences
	err := s.pool.QueryRow(ctx, `SELECT interests,activity,travel_type FROM user_preferences WHERE user_id=$1`, id).Scan(&p.Interests, &p.Activity, &p.TravelType)
	if errors.Is(err, pgx.ErrNoRows) {
		return Preferences{}, ErrNotFound
	}
	return p, err
}
func (s *PostgresStore) SavePreferences(ctx context.Context, id string, p Preferences) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO user_preferences(user_id,interests,activity,travel_type) VALUES($1,$2,$3,$4) ON CONFLICT(user_id) DO UPDATE SET interests=EXCLUDED.interests,activity=EXCLUDED.activity,travel_type=EXCLUDED.travel_type,updated_at=now()`, id, p.Interests, p.Activity, p.TravelType)
	return err
}
