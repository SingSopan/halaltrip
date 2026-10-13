package directory

import (
	"context"
	"time"
)

type Kind string

const (
	Restaurant Kind = "restaurant"
	Mosque     Kind = "mosque"
	Hotel      Kind = "hotel"
)

type HotelFeature struct {
	Name             string `json:"name"`
	Verified         bool   `json:"verified"`
	VerificationNote string `json:"verification_note,omitempty"`
}

type Place struct {
	ID                     string         `json:"id,omitempty"`
	Name                   string         `json:"name"`
	Description            string         `json:"description,omitempty"`
	City                   string         `json:"city"`
	Country                string         `json:"country"`
	Address                string         `json:"address,omitempty"`
	ImageURL               string         `json:"image_url,omitempty"`
	Latitude               *float64       `json:"latitude,omitempty"`
	Longitude              *float64       `json:"longitude,omitempty"`
	Cuisine                string         `json:"cuisine,omitempty"`
	PriceRange             string         `json:"price_range,omitempty"`
	HalalStatus            string         `json:"halal_status,omitempty"`
	HalalVerificationNote  string         `json:"halal_verification_note,omitempty"`
	Facilities             []string       `json:"facilities,omitempty"`
	MuslimFriendlyFeatures []HotelFeature `json:"muslim_friendly_features,omitempty"`
	SourceURL              string         `json:"source_url,omitempty"`
	CreatedAt              time.Time      `json:"created_at,omitempty"`
	UpdatedAt              time.Time      `json:"updated_at,omitempty"`
}

type Filters struct {
	Query       string
	City        string
	Country     string
	Cuisine     string
	HalalStatus string
	PriceRange  string
	Facility    string
	Feature     string
	Sort        string
	Order       string
	Limit       int
	Offset      int
}

type Store interface {
	List(context.Context, Kind, Filters) ([]Place, int, error)
	Get(context.Context, Kind, string) (Place, error)
	Create(context.Context, Kind, Place) (Place, error)
	Update(context.Context, Kind, string, Place) (Place, error)
	Delete(context.Context, Kind, string) error
	Save(context.Context, string, Kind, string) error
	Saved(context.Context, string) ([]SavedRef, error)
	Unsave(context.Context, string, Kind, string) error
}

type SavedRef struct {
	Kind      Kind      `json:"place_type"`
	ID        string    `json:"place_id"`
	CreatedAt time.Time `json:"created_at"`
}
