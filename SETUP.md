# MapUp Geofence Monitor - Setup Guide

This document describes how to run the MapUp Geofence Monitor locally or deploy it to production.

## Local Development (Docker Compose)

### Prerequisites

- Docker (latest stable)
- Docker Compose (usually bundled with Docker)

### Run the Stack

1. Ensure you are in the project root directory (where this README.md lives).
2. Run:
   ```bash
   docker compose up --build
   ```

This will start three services:

- **db** – PostGIS database with the MapUp schema (initialized from `db/init.sql`)
- **api** – Go API server on `http://localhost:8080`
- **frontend** – React app on `http://localhost:5173`

The database is mapped to a host port (`5433`) to avoid conflicts with an existing local PostgreSQL service.

### Verify Everything is Working

Open your browser to `http://localhost:5173` to see the control dashboard.

You can also query the API directly:

```bash
# List all geofences
curl http://localhost:8080/api/geofences

# Create a geofence
curl -X POST http://localhost:8080/api/geofences \
  -H "Content-Type: application/json" \
  -d '{"name":"Office Park","category":"restricted","polygon":[{"latitude":40.7128,"longitude":-74.0060},{"latitude":40.7138,"longitude":-74.0070},{"latitude":40.7128,"longitude":-74.0070},{"latitude":40.7128,"longitude":-74.0060}]}'
```

## Manual Build (for Development)

If you prefer not to use Docker, you can run the backend and frontend manually:

### Backend (Go)

1. Install Go 1.22 or newer.
2. Run `go build ./cmd/api` to create the binary.
3. Set environment variables:
   ```bash
   export PORT=8080
   export DATABASE_URL=postgres://mapup:mapup@localhost:5432/mapup?sslmode=disable
   export ALLOWED_ORIGIN=http://localhost:5173
   ```
4. Start PostgreSQL with PostGIS extension enabled.
5. Load the schema:
   ```bash
   psql $DATABASE_URL < db/init.sql
   ```
6. Run the binary: `./mapup-api`

### Frontend (React)

1. Install Node.js 20+.
2. `cd frontend` and run `npm install`.
3. Set environment variable:
   ```bash
   export VITE_API_URL=http://localhost:8080
   ```
4. Run `npm run dev` (or `npm run build` + `npm run preview` for production).

## Production Deployment

### Deploying the Backend

The backend is built as a Docker image and can be deployed to any container platform (Kubernetes, Docker Swarm, cloud VMs, etc.). The Docker image is pushed to Docker Hub under the name `lakshmikanthtammisetty/mapup-api`.

#### Build and Push

```bash
# Build the image locally
docker build -t lakshmikanthtammisetty/mapup-api ./backend

# Push to Docker Hub (requires authentication)
docker push lakshmikanthtammisetty/mapup-api
```

#### Deploy with Docker Compose (production)

Create a `docker-compose.prod.yml` (or extend the existing one) with:

```yaml
db:
  image: postgis/postgis:16-3.4
  environment:
    POSTGRES_USER: mapup
    POSTGRES_PASSWORD: ${DB_PASSWORD}
    POSTGRES_DB: mapup
  volumes:
    - db-data:/var/lib/postgresql/data
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U mapup -d mapup"]
    interval: 5s
    timeout: 5s
    retries: 10

api:
  image: lakshmikanthtammisetty/mapup-api:latest
  environment:
    DATABASE_URL: postgres://mapup:${DB_PASSWORD}@db:5432/mapup?sslmode=disable
    ALLOWED_ORIGIN: https://mapup.example.com
    PORT: 8080
  depends_on: [db]
  restart: unless-stopped

frontend:
  image: lakshmikanthtammisetty/mapup-frontend:latest
  environment:
    VITE_API_URL: http://localhost:8080
  depends_on: [api]
  restart: unless-stopped

volumes:
  db-data:
```

### Deploying the Frontend

The frontend is also packaged as a Docker image (`lakshmikanthtammisetty/mapup-frontend`) and can be served by any web server that supports HTTP/HTTPS. Use an HTTPS-terminating reverse proxy (nginx, Traefik) in front of the frontend container.

## Configuration

All services use environment variables:

### Backend

- `PORT` – HTTP port (default `8080`)
- `DATABASE_URL` – PostgreSQL connection string (default `postgres://mapup:mapup@localhost:5432/mapup?sslmode=disable`)
- `ALLOWED_ORIGIN` – CORS origin for the API (default `http://localhost:5173`)

### Frontend

- `VITE_API_URL` – Base URL of the backend (default `http://localhost:8080`)

## Troubleshooting

### Database Connection Issues

If the API fails to start with “could not list geofences”, ensure:

1. PostgreSQL/PostGIS is running.
2. The `mapup` user and `mapup` database exist.
3. The connection string matches the actual server.

You can verify the database is reachable:

```bash
docker compose exec db psql -U mapup -d mapup -c "SELECT PostGIS_Version();"
```

### Frontend Cannot Reach API

Check that the frontend environment variable `VITE_API_URL` points to the correct backend address and that CORS is configured in the API (`ALLOWED_ORIGIN`).

### “vehicle not found” Error

When registering a location, the API validates that the vehicle exists before processing. This prevents accidental updates for stale vehicle IDs.

## Health Checks

The database service exposes a healthcheck endpoint that the API depends on. If the database becomes unhealthy, Docker Compose will restart the API.

## Development Tips

- Use `docker compose logs -f` to see real‑time logs from all services.
- To clear the database while keeping data (e.g., for testing), remove the volume:
  ```bash
  docker volume rm mapup-mapup-data
  ```
  Docker Compose will recreate it on the next run.
- The frontend exposes hot‑reload on port 5173 when running `npm run dev` locally (outside Docker).

## Notes on Code

### Geofence Definition and Storage

The API accepts coordinates as `[latitude, longitude]` in JSON. They are stored as PostGIS geometry in **EPSG:4326** order `[longitude, latitude]`. This matches the GeoJSON spec.

### Spatial Query (`ST_Covers`)

The containment check uses `ST_Covers`. This function is **inclusive**: a point that lies exactly on a geofence edge is considered “inside” the geofence. The code includes a comment referencing this behavior.

### Entry/Exit Detection

When a vehicle updates its location, the system computes the set of geofences it is currently inside. By comparing this with the previous set, only *changes* trigger alerts – not the continuous presence inside a geofence.

### Real‑time Alerts

Alert rules are evaluated in a goroutine for each entry or exit. The alert is persisted and then broadcast over a WebSocket hub to all connected front‑end clients. The hub protects its client list with a read/write mutex and adds a write‑mutex to each client to avoid concurrent writes from corrupting the stream.

### Alert History

A `/api/alerts` endpoint returns the most recent alerts, with optional filters:

- `vehicle_id` – restrict to a specific vehicle
- `geofence_id` – restrict to a specific geofence
- `event_type` – `entry`, `exit`, or `both`
- `since` – ISO‑8601 timestamp (e.g. `2024‑12‑01T10:00:00Z`)

The `limit` query parameter caps the response size (default 100, max 1000). Filtered results preserve the same pagination and order.

## Security Considerations (Development)

This setup is intended for local development and demonstration. For production:

- Use TLS/HTTPS for both API and frontend.
- Enforce authentication (e.g., API keys, JWT) on the backend.
- Restrict database access to the API container only.
- Configure proper CORS (`ALLOWED_ORIGIN`) in production.
- Run the containers under dedicated non‑root users.
- Regularly update base images and dependencies.

## License

The MapUp Geofence Monitor is released under the MIT License. See LICENSE for details.
