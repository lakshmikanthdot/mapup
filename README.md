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
- `GET /api/alerts` — supports `?vehicle_id=`, `?geofence_id=`, `?event_type=` and `?since=` filters
- `POST /api/alert-rules`
- `GET /ws`

The first location update establishes the vehicle's inside set and emits matching entry alerts. Later updates compare the previous set and emit only changes. Alerts are persisted and broadcast asynchronously over WebSocket, so the alert pipeline never blocks the location-update response.

A location update for an unknown vehicle returns `404 vehicle not found`. Geofences use `ST_Covers`, so a point exactly on a boundary counts as inside.

## Demo

A 30–45 minute demo video covering geofence creation, vehicle entry/exit, live WebSocket alerts, alert history filters and Docker running locally:

**[Demo Video]** — _(link to be added after recording)_

## Deployment

The backend and frontend are each built as Docker images using multi-stage builds:

- Backend: built from `./backend`, published to Docker Hub as `lakshmikanthtammisetty/mapup-api`
- Frontend: built from `./frontend`, published as `lakshmikanthtammisetty/mapup-frontend`

See [SETUP.md](./SETUP.md) for full local setup and production deployment instructions.

Repo: https://github.com/lakshmikanthdot/mapup
