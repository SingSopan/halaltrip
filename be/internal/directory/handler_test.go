package directory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"halaltrip/be/internal/auth"
	"halaltrip/be/internal/middleware"
)

type fakeStore struct {
	place       Place
	filters     Filters
	listedKind  Kind
	listUser    string
	savedUser   string
	createdKind Kind
}

func (s *fakeStore) List(_ context.Context, k Kind, f Filters) ([]Place, int, error) {
	s.filters = f
	s.listedKind = k
	return []Place{s.place}, 1, nil
}
func (s *fakeStore) Get(_ context.Context, _ Kind, id string) (Place, error) {
	p := s.place
	p.ID = id
	return p, nil
}
func (s *fakeStore) Create(_ context.Context, k Kind, p Place) (Place, error) {
	s.createdKind = k
	s.place = p
	p.ID = "123e4567-e89b-42d3-a456-426614174000"
	return p, nil
}
func (s *fakeStore) Update(_ context.Context, k Kind, id string, p Place) (Place, error) {
	p.ID = id
	return p, nil
}
func (s *fakeStore) Delete(context.Context, Kind, string) error { return nil }
func (s *fakeStore) Save(_ context.Context, user string, k Kind, id string) error {
	s.savedUser = user
	return nil
}
func (s *fakeStore) Saved(_ context.Context, user string) ([]SavedRef, error) {
	s.listUser = user
	return []SavedRef{{Kind: Restaurant, ID: "123e4567-e89b-42d3-a456-426614174000", CreatedAt: time.Now()}}, nil
}
func (s *fakeStore) Unsave(context.Context, string, Kind, string) error { return nil }

var testSecret = []byte("directory-test-jwt-secret-at-least-32-bytes")

func directoryRouter(store Store) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(store)
	r := gin.New()
	r.GET("/restaurants", h.List(Restaurant))
	r.GET("/restaurants/:id", h.Get(Restaurant))
	r.POST("/admin/restaurants", h.Create(Restaurant))
	r.POST("/admin/mosques", h.Create(Mosque))
	r.POST("/admin/hotels", h.Create(Hotel))
	protected := r.Group("", middleware.JWT(testSecret, "", "", false))
	protected.POST("/saved-places", h.Save)
	protected.GET("/saved-places", h.ListSaved)
	return r
}
func request(r http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	r.ServeHTTP(w, req)
	return w
}
func userToken(t *testing.T, id string) string {
	t.Helper()
	token, err := auth.SignToken(testSecret, id, "", "", time.Hour, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestListPaginationAndFilters(t *testing.T) {
	store := &fakeStore{}
	res := request(directoryRouter(store), http.MethodGet, "/restaurants?q=spice&city=Jakarta&country=Indonesia&halal_status=unknown&page=2&page_size=5&sort=city&order=desc", "", "")
	if res.Code != 200 {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if store.listedKind != Restaurant || store.filters.Query != "spice" || store.filters.City != "Jakarta" || store.filters.Country != "Indonesia" || store.filters.HalalStatus != "unknown" || store.filters.Offset != 5 || store.filters.Limit != 5 || store.filters.Sort != "city" || store.filters.Order != "desc" {
		t.Fatalf("filters were not passed correctly: %+v", store.filters)
	}
	if bad := request(directoryRouter(store), http.MethodGet, "/restaurants?sort=name%20DESC%3B%20DROP%20TABLE%20users", "", ""); bad.Code != 400 {
		t.Fatalf("unsafe sort status=%d", bad.Code)
	}
}
func TestPlaceValidationRequiresCoordinatesAndEvidence(t *testing.T) {
	lat, lon := 91.0, 10.0
	if err := ValidatePlace(Restaurant, Place{Name: "Cafe", City: "Jakarta", Country: "Indonesia", Latitude: &lat, Longitude: &lon}); err == nil {
		t.Fatal("accepted out-of-range coordinates")
	}
	if err := ValidatePlace(Restaurant, Place{Name: "Cafe", City: "Jakarta", Country: "Indonesia", HalalStatus: "certified"}); err == nil {
		t.Fatal("accepted unsupported certification claim")
	}
	if err := ValidatePlace(Hotel, Place{Name: "Hotel", City: "Jakarta", Country: "Indonesia", MuslimFriendlyFeatures: []HotelFeature{{Name: "Prayer mat", Verified: true}}}); err == nil {
		t.Fatal("accepted verified feature without evidence")
	}
}
func TestAdminCreateNormalizesUnknownHalalStatusAndMosqueFacilities(t *testing.T) {
	store := &fakeStore{}
	r := directoryRouter(store)
	res := request(r, http.MethodPost, "/admin/restaurants", `{"name":"Cafe","city":"Jakarta","country":"Indonesia"}`, "")
	if res.Code != http.StatusCreated || store.place.HalalStatus != "unknown" {
		t.Fatalf("create=%d body=%s status=%q", res.Code, res.Body.String(), store.place.HalalStatus)
	}
	res = request(r, http.MethodPost, "/admin/mosques", `{"name":"Masjid","city":"Jakarta","country":"Indonesia","facilities":["women_prayer_area"]}`, "")
	if res.Code != http.StatusCreated || store.createdKind != Mosque || len(store.place.Facilities) != 1 {
		t.Fatalf("mosque create=%d %s", res.Code, res.Body.String())
	}
}
func TestSavedPlacesUseAuthenticatedUserSubject(t *testing.T) {
	store := &fakeStore{}
	r := directoryRouter(store)
	token := userToken(t, "user-from-token")
	body := `{"place_type":"restaurant","place_id":"123e4567-e89b-42d3-a456-426614174000"}`
	if res := request(r, http.MethodPost, "/saved-places", body, ""); res.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous save status=%d", res.Code)
	}
	if res := request(r, http.MethodPost, "/saved-places", body, token); res.Code != http.StatusCreated || store.savedUser != "user-from-token" {
		t.Fatalf("save status=%d user=%q", res.Code, store.savedUser)
	}
	if res := request(r, http.MethodGet, "/saved-places", "", token); res.Code != http.StatusOK || store.listUser != "user-from-token" {
		t.Fatalf("list saved status=%d user=%q", res.Code, store.listUser)
	}
}
