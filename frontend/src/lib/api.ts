export type Coordinate = { latitude: number; longitude: number };
export type Geofence = { id: string; name: string; category: string; polygon: Coordinate[] };
export type Vehicle = { id: string; name: string };
export type Alert = { id: string; vehicle_id: string; geofence_id: string; event_type: string; message: string; created_at: string };
const API = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API}${path}`, { headers: {'Content-Type': 'application/json'}, ...init });
  const body = await res.json();
  if (!res.ok) throw new Error(body.data?.error ?? 'Request failed');
  return body.data;
}
export const api = {
  geofences: () => request<Geofence[]>('/api/geofences'),
  vehicles: () => request<Vehicle[]>('/api/vehicles'),
  alerts: () => request<Alert[]>('/api/alerts'),
  createGeofence: (body: Omit<Geofence, 'id'>) => request<Geofence>('/api/geofences', {method:'POST', body:JSON.stringify(body)}),
  createVehicle: (name: string) => request<Vehicle>('/api/vehicles', {method:'POST', body:JSON.stringify({name})}),
  createRule: (geofence_id: string, event_type: string) => request('/api/alert-rules', {method:'POST', body:JSON.stringify({geofence_id,event_type})}),
  updateLocation: (id: string, body: Coordinate) => request(`/api/vehicles/${id}/location`, {method:'POST', body:JSON.stringify(body)}),
  websocket: () => `${API.replace(/^http/, 'ws')}/ws`
};
