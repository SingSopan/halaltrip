package integration

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"halaltrip/be/internal/directory"
	"halaltrip/be/internal/trips"
	"halaltrip/be/internal/users"
)

func TestPostgresStoresAgainstMigrations(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgxpool.NewWithConfig(ctx, adminCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if err := admin.Ping(ctx); err != nil {
		t.Fatalf("connect TEST_DATABASE_URL: %v", err)
	}
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatal(err)
	}
	schema := "halaltrip_test_" + hex.EncodeToString(random[:])
	if _, err := admin.Exec(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(context.Background(), `DROP SCHEMA "`+schema+`" CASCADE`) }()

	appCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if appCfg.ConnConfig.RuntimeParams == nil {
		appCfg.ConnConfig.RuntimeParams = make(map[string]string)
	}
	appCfg.ConnConfig.RuntimeParams["search_path"] = schema
	app, err := pgxpool.NewWithConfig(ctx, appCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	applyMigrations(t, ctx, app)

	userStore := users.NewPostgresStore(app)
	owner, err := userStore.Create(ctx, "Integration Owner", "owner@example.test", "not-a-real-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	other, err := userStore.Create(ctx, "Other User", "other@example.test", "not-a-real-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	if err := userStore.SavePreferences(ctx, owner.ID, users.Preferences{Interests: []string{"culture"}, Activity: "Exploration", TravelType: "Solo"}); err != nil {
		t.Fatal(err)
	}
	if got, err := userStore.GetPreferences(ctx, owner.ID); err != nil || got.Activity != "Exploration" {
		t.Fatalf("GetPreferences() = %+v, %v", got, err)
	}

	tripStore := trips.NewPostgresStore(app)
	trip, err := tripStore.CreateTrip(ctx, owner.ID, trips.Trip{Title: "Integration Trip", StartDate: "2030-05-01", EndDate: "2030-05-03", Status: trips.StatusPlanning})
	if err != nil {
		t.Fatal(err)
	}
	first, err := tripStore.CreateItem(ctx, owner.ID, trip.ID, trips.Item{Title: "Visit mosque", PlaceType: "mosque"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := tripStore.CreateItem(ctx, owner.ID, trip.ID, trips.Item{Title: "Dinner", PlaceType: "restaurant"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tripStore.GetTrip(ctx, other.ID, trip.ID); err != trips.ErrNotFound {
		t.Fatalf("other account GetTrip error = %v, want ErrNotFound", err)
	}
	if err := tripStore.ReorderItems(ctx, owner.ID, trip.ID, []string{second.ID, first.ID}); err != nil {
		t.Fatal(err)
	}
	items, err := tripStore.ListItems(ctx, owner.ID, trip.ID)
	if err != nil || len(items) != 2 || items[0].ID != second.ID || items[0].Sequence != 1 {
		t.Fatalf("reordered items = %+v, %v", items, err)
	}
	if err := tripStore.DeleteItem(ctx, owner.ID, trip.ID, second.ID); err != nil {
		t.Fatal(err)
	}
	items, err = tripStore.ListItems(ctx, owner.ID, trip.ID)
	if err != nil || len(items) != 1 || items[0].ID != first.ID || items[0].Sequence != 1 {
		t.Fatalf("items after delete = %+v, %v", items, err)
	}

	placeStore := directory.NewPostgresStore(app)
	place, err := placeStore.Create(ctx, directory.Restaurant, directory.Place{Name: "Test Halal Kitchen", City: "Jakarta", Country: "Indonesia", HalalStatus: "unknown"})
	if err != nil {
		t.Fatal(err)
	}
	if err := placeStore.Save(ctx, owner.ID, directory.Restaurant, place.ID); err != nil {
		t.Fatal(err)
	}
	if err := placeStore.Save(ctx, owner.ID, directory.Restaurant, place.ID); err != nil {
		t.Fatalf("duplicate save should be idempotent: %v", err)
	}
	saved, err := placeStore.Saved(ctx, owner.ID)
	if err != nil || len(saved) != 1 || saved[0].ID != place.ID {
		t.Fatalf("saved refs = %+v, %v", saved, err)
	}
}

func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	migrationDir := filepath.Join(filepath.Dir(file), "..", "migrations")
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		contents, err := os.ReadFile(filepath.Join(migrationDir, name))
		if err != nil {
			t.Fatal(err)
		}
		up, _, _ := strings.Cut(string(contents), "-- +goose Down")
		up = strings.TrimPrefix(up, "-- +goose Up")
		for _, statement := range strings.Split(up, ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if _, err := pool.Exec(ctx, statement); err != nil {
				t.Fatalf("apply %s: %v", name, err)
			}
		}
	}
}
