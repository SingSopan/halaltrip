package trips

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"halaltrip/be/internal/auth"
	"halaltrip/be/internal/middleware"
)

const (
	tripID  = "123e4567-e89b-42d3-a456-426614174000"
	tripNew = "123e4567-e89b-42d3-a456-426614174010"
	itemA   = "123e4567-e89b-42d3-a456-426614174001"
	itemB   = "123e4567-e89b-42d3-a456-426614174002"
)

var tripTestSecret = []byte("trip-planner-tests-jwt-secret-at-least-32-bytes")

type fakeTripStore struct {
	trips     map[string]Trip
	owners    map[string]string
	items     map[string][]Item
	reorders  int
	lastOrder []string
}

func newFakeTripStore() *fakeTripStore {
	return &fakeTripStore{trips: map[string]Trip{tripID: {ID: tripID, Title: "Japan", StartDate: "2026-11-01", EndDate: "2026-11-03", Status: StatusPlanning}}, owners: map[string]string{tripID: "owner-a"}, items: map[string][]Item{}}
}
func (s *fakeTripStore) owns(user, id string) bool { return s.owners[id] == user }
func (s *fakeTripStore) CreateTrip(_ context.Context, user string, t Trip) (Trip, error) {
	t.ID = tripNew
	if _, ok := s.trips[t.ID]; ok {
		return Trip{}, errors.New("duplicate")
	}
	s.trips[t.ID] = t
	s.owners[t.ID] = user
	return t, nil
}
func (s *fakeTripStore) ListTrips(_ context.Context, user string) ([]Trip, error) {
	out := []Trip{}
	for id, t := range s.trips {
		if s.owns(user, id) {
			out = append(out, t)
		}
	}
	return out, nil
}
func (s *fakeTripStore) GetTrip(_ context.Context, user, id string) (Trip, error) {
	if !s.owns(user, id) {
		return Trip{}, ErrNotFound
	}
	return s.trips[id], nil
}
func (s *fakeTripStore) UpdateTrip(_ context.Context, user, id string, p TripPatch) (Trip, error) {
	t, err := s.GetTrip(context.Background(), user, id)
	if err != nil {
		return Trip{}, err
	}
	if p.Title != nil {
		t.Title = *p.Title
	}
	if p.StartDate != nil {
		t.StartDate = *p.StartDate
	}
	if p.EndDate != nil {
		t.EndDate = *p.EndDate
	}
	if p.Origin != nil {
		t.Origin = *p.Origin
	}
	if p.Notes != nil {
		t.Notes = *p.Notes
	}
	if p.Status != nil {
		t.Status = *p.Status
	}
	s.trips[id] = t
	return t, nil
}
func (s *fakeTripStore) DeleteTrip(_ context.Context, user, id string) error {
	if !s.owns(user, id) {
		return ErrNotFound
	}
	delete(s.trips, id)
	delete(s.owners, id)
	delete(s.items, id)
	return nil
}
func (s *fakeTripStore) CreateItem(_ context.Context, user, trip string, i Item) (Item, error) {
	if !s.owns(user, trip) {
		return Item{}, ErrNotFound
	}
	if len(s.items[trip]) == 0 {
		i.ID = itemA
	} else {
		i.ID = itemB
	}
	i.TripID = trip
	i.Sequence = len(s.items[trip]) + 1
	s.items[trip] = append(s.items[trip], i)
	return i, nil
}
func (s *fakeTripStore) ListItems(_ context.Context, user, trip string) ([]Item, error) {
	if !s.owns(user, trip) {
		return nil, ErrNotFound
	}
	return append([]Item{}, s.items[trip]...), nil
}
func (s *fakeTripStore) UpdateItem(_ context.Context, user, trip, id string, p ItemPatch) (Item, error) {
	if !s.owns(user, trip) {
		return Item{}, ErrNotFound
	}
	for n, i := range s.items[trip] {
		if i.ID == id {
			v := merge(i, p)
			s.items[trip][n] = v
			return v, nil
		}
	}
	return Item{}, ErrNotFound
}
func (s *fakeTripStore) DeleteItem(_ context.Context, user, trip, id string) error {
	if !s.owns(user, trip) {
		return ErrNotFound
	}
	items := s.items[trip]
	for n, i := range items {
		if i.ID == id {
			s.items[trip] = append(items[:n], items[n+1:]...)
			return nil
		}
	}
	return ErrNotFound
}
func (s *fakeTripStore) ReorderItems(_ context.Context, user, trip string, ids []string) error {
	if !s.owns(user, trip) {
		return ErrNotFound
	}
	items := s.items[trip]
	if len(items) != len(ids) {
		return ErrOrderChanged
	}
	byID := map[string]Item{}
	for _, i := range items {
		byID[i.ID] = i
	}
	for n, id := range ids {
		i, ok := byID[id]
		if !ok {
			return ErrOrderChanged
		}
		i.Sequence = n + 1
		s.items[trip][n] = i
	}
	s.reorders++
	s.lastOrder = append([]string{}, ids...)
	return nil
}

func testRouter(store Store) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(store)
	r := gin.New()
	g := r.Group("/api/v1/trips", middleware.JWT(tripTestSecret, "", "", false))
	g.POST("", h.CreateTrip)
	g.GET("", h.ListTrips)
	g.GET("/:tripId", h.GetTrip)
	g.PATCH("/:tripId", h.PatchTrip)
	g.DELETE("/:tripId", h.DeleteTrip)
	g.POST("/:tripId/itinerary-items", h.CreateItem)
	g.GET("/:tripId/itinerary-items", h.ListItems)
	g.PUT("/:tripId/itinerary-items/order", h.ReorderItems)
	g.PATCH("/:tripId/itinerary-items/:itemId", h.PatchItem)
	g.DELETE("/:tripId/itinerary-items/:itemId", h.DeleteItem)
	return r
}
func signedToken(t *testing.T, user string) string {
	t.Helper()
	v, err := auth.SignToken(tripTestSecret, user, "", "", time.Hour, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func perform(r http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestTripCRUDAndOwnerIsolation(t *testing.T) {
	store := newFakeTripStore()
	r := testRouter(store)
	owner := signedToken(t, "owner-a")
	other := signedToken(t, "owner-b")
	if res := perform(r, http.MethodPost, "/api/v1/trips", `{"title":"Trip","start_date":"2026-11-01","end_date":"2026-11-03"}`, ""); res.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous create status=%d", res.Code)
	}
	created := perform(r, http.MethodPost, "/api/v1/trips", `{"title":"Trip","start_date":"2026-11-01","end_date":"2026-11-03","user_id":"owner-b"}`, owner)
	if created.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", created.Code, created.Body.String())
	}
	if got := perform(r, http.MethodGet, "/api/v1/trips/"+tripID, "", owner); got.Code != 200 {
		t.Fatalf("owner get=%d", got.Code)
	}
	if got := perform(r, http.MethodGet, "/api/v1/trips", "", other); got.Code != 200 || strings.Contains(got.Body.String(), tripID) {
		t.Fatalf("cross-user trip list=%d %s", got.Code, got.Body.String())
	}
	if got := perform(r, http.MethodGet, "/api/v1/trips/"+tripID, "", other); got.Code != 404 {
		t.Fatalf("cross-user get=%d", got.Code)
	}
	patched := perform(r, http.MethodPatch, "/api/v1/trips/"+tripID, `{"title":"Updated","status":"active"}`, owner)
	if patched.Code != 200 || store.trips[tripID].Title != "Updated" {
		t.Fatalf("patch=%d %s", patched.Code, patched.Body.String())
	}
	if got := perform(r, http.MethodDelete, "/api/v1/trips/"+tripID, "", other); got.Code != 404 {
		t.Fatalf("cross-user delete=%d", got.Code)
	}
	if got := perform(r, http.MethodDelete, "/api/v1/trips/"+tripID, "", owner); got.Code != 204 {
		t.Fatalf("delete=%d", got.Code)
	}
}

func TestTripAndItineraryDateValidation(t *testing.T) {
	store := newFakeTripStore()
	r := testRouter(store)
	token := signedToken(t, "owner-a")
	for _, body := range []string{`{"title":"Bad","start_date":"2026-11-04","end_date":"2026-11-03"}`, `{"title":"Bad","start_date":"11/01/2026","end_date":"2026-11-03"}`} {
		if res := perform(r, http.MethodPost, "/api/v1/trips", body, token); res.Code != 400 {
			t.Errorf("invalid trip date status=%d body=%s", res.Code, res.Body.String())
		}
	}
	outside := perform(r, http.MethodPost, "/api/v1/trips/"+tripID+"/itinerary-items", `{"title":"Visit","place_type":"destination","scheduled_date":"2026-11-04"}`, token)
	if outside.Code != 400 {
		t.Fatalf("outside date status=%d", outside.Code)
	}
	badTime := perform(r, http.MethodPost, "/api/v1/trips/"+tripID+"/itinerary-items", `{"title":"Visit","place_type":"destination","scheduled_date":"2026-11-02","start_time":"18:00","end_time":"17:00"}`, token)
	if badTime.Code != 400 {
		t.Fatalf("bad time status=%d", badTime.Code)
	}
	valid := perform(r, http.MethodPost, "/api/v1/trips/"+tripID+"/itinerary-items", `{"title":"Dinner","place_type":"restaurant","external_place_id":"local-rest-7","scheduled_date":"2026-11-02","start_time":"18:00","end_time":"19:30"}`, token)
	if valid.Code != 201 || store.items[tripID][0].Sequence != 1 {
		t.Fatalf("valid item=%d %s", valid.Code, valid.Body.String())
	}
	list := perform(r, http.MethodGet, "/api/v1/trips/"+tripID+"/itinerary-items", "", token)
	if list.Code != 200 || !strings.Contains(list.Body.String(), "local-rest-7") {
		t.Fatalf("list items=%d %s", list.Code, list.Body.String())
	}
}

func TestItineraryCRUDSequenceAndTransactionalReorderContract(t *testing.T) {
	store := newFakeTripStore()
	store.items[tripID] = []Item{{ID: itemA, TripID: tripID, Title: "First", PlaceType: "custom_activity", Sequence: 1}, {ID: itemB, TripID: tripID, Title: "Second", PlaceType: "mosque", Sequence: 2}}
	r := testRouter(store)
	token := signedToken(t, "owner-a")
	base := "/api/v1/trips/" + tripID + "/itinerary-items"
	bad := perform(r, http.MethodPut, base+"/order", `{"item_ids":["`+itemA+`","`+itemA+`"]}`, token)
	if bad.Code != 400 || store.reorders != 0 {
		t.Fatalf("invalid order changed data: status=%d calls=%d", bad.Code, store.reorders)
	}
	good := perform(r, http.MethodPut, base+"/order", `{"item_ids":["`+itemB+`","`+itemA+`"]}`, token)
	if good.Code != 204 || store.reorders != 1 || store.items[tripID][0].ID != itemB || store.items[tripID][0].Sequence != 1 {
		t.Fatalf("reorder=%d items=%+v", good.Code, store.items[tripID])
	}
	if res := perform(r, http.MethodPatch, base+"/"+itemB, `{"title":"Updated"}`, token); res.Code != 200 || store.items[tripID][0].Title != "Updated" {
		t.Fatalf("patch item=%d %s", res.Code, res.Body.String())
	}
	if res := perform(r, http.MethodDelete, base+"/"+itemA, "", token); res.Code != 204 {
		t.Fatalf("delete item=%d", res.Code)
	}
	foreign := signedToken(t, "owner-b")
	if res := perform(r, http.MethodPatch, base+"/"+itemB, `{"title":"stolen"}`, foreign); res.Code != 404 {
		t.Fatalf("cross-user item patch=%d", res.Code)
	}
}

func TestTripPatchCannotExcludeExistingScheduledItems(t *testing.T) {
	store := newFakeTripStore()
	day := "2026-11-03"
	store.items[tripID] = []Item{{ID: itemA, TripID: tripID, Title: "Visit", PlaceType: "destination", ScheduledDate: &day, Sequence: 1}}
	r := testRouter(store)
	token := signedToken(t, "owner-a")
	res := perform(r, http.MethodPatch, "/api/v1/trips/"+tripID, `{"end_date":"2026-11-02"}`, token)
	if res.Code != 400 {
		t.Fatalf("date range patch status=%d body=%s", res.Code, res.Body.String())
	}
}
