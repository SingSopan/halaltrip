package trips

import (
	"context"
	"time"
)

const (
	StatusPlanning  = "planning"
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

type Trip struct {
	ID        string    `json:"id,omitempty"`
	Title     string    `json:"title"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Origin    string    `json:"origin,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type TripPatch struct {
	Title     *string `json:"title"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
	Origin    *string `json:"origin"`
	Notes     *string `json:"notes"`
	Status    *string `json:"status"`
}

type Item struct {
	ID              string    `json:"id,omitempty"`
	TripID          string    `json:"trip_id,omitempty"`
	Title           string    `json:"title"`
	Description     string    `json:"description,omitempty"`
	PlaceType       string    `json:"place_type"`
	ExternalPlaceID *string   `json:"external_place_id,omitempty"`
	ScheduledDate   *string   `json:"scheduled_date,omitempty"`
	StartTime       *string   `json:"start_time,omitempty"`
	EndTime         *string   `json:"end_time,omitempty"`
	Sequence        int       `json:"sequence"`
	TransportMode   string    `json:"transport_mode,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

type ItemPatch struct {
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	PlaceType       *string `json:"place_type"`
	ExternalPlaceID *string `json:"external_place_id"`
	ScheduledDate   *string `json:"scheduled_date"`
	StartTime       *string `json:"start_time"`
	EndTime         *string `json:"end_time"`
	TransportMode   *string `json:"transport_mode"`
	Notes           *string `json:"notes"`
}

type Store interface {
	CreateTrip(context.Context, string, Trip) (Trip, error)
	ListTrips(context.Context, string) ([]Trip, error)
	GetTrip(context.Context, string, string) (Trip, error)
	UpdateTrip(context.Context, string, string, TripPatch) (Trip, error)
	DeleteTrip(context.Context, string, string) error
	CreateItem(context.Context, string, string, Item) (Item, error)
	ListItems(context.Context, string, string) ([]Item, error)
	UpdateItem(context.Context, string, string, string, ItemPatch) (Item, error)
	DeleteItem(context.Context, string, string, string) error
	ReorderItems(context.Context, string, string, []string) error
}

type ErrValidation struct{ Message string }

func (e ErrValidation) Error() string { return e.Message }
