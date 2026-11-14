package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"mapup/backend/internal/models"
)

type Store struct { Pool *pgxpool.Pool }

func New(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil { return nil, err }
	if err := pool.Ping(ctx); err != nil { pool.Close(); return nil, err }
	return &Store{Pool: pool}, nil
}

func polygonWKT(points []models.Coordinate) string {
	values := make([]string, len(points))
	for i, p := range points { values[i] = fmt.Sprintf("%f %f", p.Longitude, p.Latitude) }
	return "POLYGON((" + strings.Join(values, ",") + "))"
}

func (s *Store) CreateGeofence(ctx context.Context, g models.Geofence) (models.Geofence, error) {
	err := s.Pool.QueryRow(ctx, `INSERT INTO geofences (name, category, boundary) VALUES ($1,$2,ST_GeomFromText($3,4326)) RETURNING id,created_at`, g.Name, g.Category, polygonWKT(g.Polygon)).Scan(&g.ID, &g.CreatedAt)
	return g, err
}

func (s *Store) ListGeofences(ctx context.Context) ([]models.Geofence, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id,name,category,ST_AsText(boundary),created_at FROM geofences ORDER BY created_at DESC`)
	if err != nil { return nil, err }; defer rows.Close()
	result := []models.Geofence{}
	for rows.Next() {
		var g models.Geofence; var wkt string
		if err := rows.Scan(&g.ID,&g.Name,&g.Category,&wkt,&g.CreatedAt); err != nil { return nil, err }
		g.Polygon = parseWKT(wkt); result = append(result,g)
	}
	return result, rows.Err()
}

func parseWKT(wkt string) []models.Coordinate {
	wkt = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(wkt, "POLYGON(("), "POLYGON (("), "))")
	parts := strings.Split(wkt, ","); result := make([]models.Coordinate, 0, len(parts))
	for _, part := range parts { var lng, lat float64; if _, err := fmt.Sscanf(strings.TrimSpace(part), "%f %f", &lng, &lat); err == nil { result = append(result, models.Coordinate{Latitude: lat, Longitude: lng}) } }
	return result
}

func (s *Store) CreateVehicle(ctx context.Context, v models.Vehicle) (models.Vehicle, error) {
	err := s.Pool.QueryRow(ctx, `INSERT INTO vehicles (name) VALUES ($1) RETURNING id,created_at`, v.Name).Scan(&v.ID,&v.CreatedAt)
	return v, err
}

func (s *Store) ListVehicles(ctx context.Context) ([]models.Vehicle, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id,name,created_at FROM vehicles ORDER BY created_at DESC`); if err != nil { return nil,err }; defer rows.Close()
	result := []models.Vehicle{}
	for rows.Next() { var v models.Vehicle; if err:=rows.Scan(&v.ID,&v.Name,&v.CreatedAt); err!=nil{return nil,err}; result=append(result,v) }
	return result, rows.Err()
}

func (s *Store) UpdateLocation(ctx context.Context, vehicleID string, location models.LocationUpdate) error {
	_, err := s.Pool.Exec(ctx, `INSERT INTO vehicle_locations (vehicle_id, location) VALUES ($1,ST_SetSRID(ST_MakePoint($2,$3),4326))`, vehicleID, location.Longitude, location.Latitude)
	return err
}

func (s *Store) CurrentLocation(ctx context.Context, vehicleID string) (*models.LocationUpdate, error) {
	var lng, lat float64
	err := s.Pool.QueryRow(ctx, `SELECT ST_X(location),ST_Y(location) FROM vehicle_locations WHERE vehicle_id=$1 ORDER BY recorded_at DESC LIMIT 1`, vehicleID).Scan(&lng,&lat)
	if err != nil { return nil, err }; return &models.LocationUpdate{Latitude:lat,Longitude:lng},nil
}

func (s *Store) GeofencesContaining(ctx context.Context, location models.LocationUpdate) ([]models.Geofence, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id,name,category,created_at FROM geofences WHERE ST_Covers(boundary,ST_SetSRID(ST_MakePoint($1,$2),4326))`, location.Longitude, location.Latitude)
	if err != nil { return nil,err }; defer rows.Close(); result:=[]models.Geofence{}
	for rows.Next(){var g models.Geofence; if err:=rows.Scan(&g.ID,&g.Name,&g.Category,&g.CreatedAt);err!=nil{return nil,err};result=append(result,g)}
	return result,rows.Err()
}

func (s *Store) CreateAlert(ctx context.Context, a models.Alert) (models.Alert,error) {
	err:=s.Pool.QueryRow(ctx,`INSERT INTO alerts(vehicle_id,geofence_id,event_type,message) VALUES($1,$2,$3,$4) RETURNING id,created_at`,a.VehicleID,a.GeofenceID,a.EventType,a.Message).Scan(&a.ID,&a.CreatedAt);return a,err
}

func (s *Store) ListAlerts(ctx context.Context, limit int) ([]models.Alert,error) {
	rows,err:=s.Pool.Query(ctx,`SELECT id,vehicle_id,geofence_id,event_type,message,created_at FROM alerts ORDER BY created_at DESC LIMIT $1`,limit);if err!=nil{return nil,err};defer rows.Close();result:=[]models.Alert{}
	for rows.Next(){var a models.Alert;if err:=rows.Scan(&a.ID,&a.VehicleID,&a.GeofenceID,&a.EventType,&a.Message,&a.CreatedAt);err!=nil{return nil,err};result=append(result,a)};return result,rows.Err()
}

func (s *Store) MatchingRules(ctx context.Context, vehicleID, geofenceID, eventType string) ([]models.AlertRule,error) {
	rows,err:=s.Pool.Query(ctx,`SELECT id,vehicle_id,geofence_id,event_type,enabled FROM alert_rules WHERE geofence_id=$1 AND event_type IN ($2,'both') AND enabled AND (vehicle_id=$3 OR vehicle_id IS NULL)`,geofenceID,eventType,vehicleID);if err!=nil{return nil,err};defer rows.Close();result:=[]models.AlertRule{}
	for rows.Next(){var r models.AlertRule;if err:=rows.Scan(&r.ID,&r.VehicleID,&r.GeofenceID,&r.EventType,&r.Enabled);err!=nil{return nil,err};result=append(result,r)};return result,rows.Err()
}

func (s *Store) CreateRule(ctx context.Context,r models.AlertRule)(models.AlertRule,error){err:=s.Pool.QueryRow(ctx,`INSERT INTO alert_rules(vehicle_id,geofence_id,event_type) VALUES($1,$2,$3) RETURNING id,enabled`,r.VehicleID,r.GeofenceID,r.EventType).Scan(&r.ID,&r.Enabled);return r,err}
