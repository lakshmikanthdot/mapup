package models

import "time"

var Categories = map[string]bool{"restricted": true, "commercial": true, "residential": true, "industrial": true}

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Geofence struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Category  string       `json:"category"`
	Polygon   []Coordinate `json:"polygon"`
	CreatedAt time.Time    `json:"created_at"`
}

type Vehicle struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type LocationUpdate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type AlertRule struct {
	ID         string `json:"id"`
	VehicleID  *string `json:"vehicle_id,omitempty"`
	GeofenceID string `json:"geofence_id"`
	EventType  string `json:"event_type"`
	Enabled    bool   `json:"enabled"`
}

type Alert struct {
	ID         string    `json:"id"`
	VehicleID  string    `json:"vehicle_id"`
	GeofenceID string    `json:"geofence_id"`
	EventType  string    `json:"event_type"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}
