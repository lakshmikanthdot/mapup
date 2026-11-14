package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"mapup/backend/internal/db"
	"mapup/backend/internal/models"
	"mapup/backend/internal/realtime"
)

type GeofenceService struct { store *db.Store; hub *realtime.Hub; mu sync.Mutex; inside map[string]map[string]bool }
func New(store *db.Store,hub *realtime.Hub)*GeofenceService{return &GeofenceService{store:store,hub:hub,inside:make(map[string]map[string]bool)}}
func(s *GeofenceService) ProcessLocation(ctx context.Context,vehicleID string,location models.LocationUpdate) error {
	current,err:=s.store.GeofencesContaining(ctx,location);if err!=nil{return err}
	next:=make(map[string]bool,len(current));for _,g:=range current{next[g.ID]=true}
	s.mu.Lock();previous,known:=s.inside[vehicleID];if !known{previous=map[string]bool{}};s.inside[vehicleID]=next;s.mu.Unlock()
	for _,g:=range current{if !previous[g.ID]{go s.emit(context.Background(),vehicleID,g,"entry")}}
	for id:=range previous{if !next[id]{go s.emit(context.Background(),vehicleID,models.Geofence{ID:id},"exit")}}
	return s.store.UpdateLocation(ctx,vehicleID,location)
}
func(s *GeofenceService) emit(ctx context.Context,vehicleID string,g models.Geofence,event string){rules,err:=s.store.MatchingRules(ctx,vehicleID,g.ID,event);if err!=nil{log.Printf("matching alert rules: %v",err);return};if len(rules)==0{return};a:=models.Alert{VehicleID:vehicleID,GeofenceID:g.ID,EventType:event,Message:fmt.Sprintf("Vehicle %s %s geofence %s",vehicleID,event,g.ID)};created,err:=s.store.CreateAlert(ctx,a);if err!=nil{log.Printf("persisting alert: %v",err);return};s.hub.Broadcast(created)}
