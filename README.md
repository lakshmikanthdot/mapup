# MapUp Geofence Monitor

A full-stack geofencing and real-time alert demo. The API is Go + PostGIS; the dashboard is React + Leaflet.

## Run locally

1. Start the complete stack: `docker compose up --build`
2. Open http://localhost:5173

The database is exposed on host port `5433` to avoid conflicts with an existing local PostgreSQL service.

Draw a closed polygon by clicking on the map, save it, register a vehicle, and create an alert rule with:

```bash
curl -X POST http://localhost:8080/api/alert-rules \
  -H 'Content-Type: application/json' \
  -d '{"geofence_id":"<id>","event_type":"both"}'
```

Coordinates accepted by the API are `[latitude, longitude]` in JSON. They are converted to PostGIS/GeoJSON order `[longitude, latitude]` at the persistence boundary. Every API response includes `time_ns`, including validation errors.

## API

- `GET|POST /api/geofences`
- `GET|POST /api/vehicles`
- `POST /api/vehicles/:id/location`
- `GET /api/alerts`
- `POST /api/alert-rules`
- `GET /ws`

The first location update establishes the vehicle's inside set and emits matching entry alerts. Later updates compare the previous set and emit only changes. Alerts are persisted and broadcast asynchronously over WebSocket.
