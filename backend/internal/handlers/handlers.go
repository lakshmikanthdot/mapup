package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"mapup/backend/internal/db"
	"mapup/backend/internal/models"
	"mapup/backend/internal/realtime"
	"mapup/backend/internal/service"
)

type API struct { Store *db.Store; Geo *service.GeofenceService; Hub *realtime.Hub; Upgrader websocket.Upgrader }
func New(store *db.Store,geo *service.GeofenceService,hub *realtime.Hub)*API{return &API{Store:store,Geo:geo,Hub:hub,Upgrader:websocket.Upgrader{CheckOrigin:func(*http.Request)bool{return true}}}}
func(a *API) respond(w http.ResponseWriter,status int,data any,start time.Time){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(map[string]any{"data":data,"time_ns":time.Since(start).Nanoseconds()})}
func(a *API) error(w http.ResponseWriter,status int,message string,start time.Time){a.respond(w,status,map[string]string{"error":message},start)}
func decode(r *http.Request,v any)error{return json.NewDecoder(r.Body).Decode(v)}
func validCoordinate(c models.Coordinate)bool{return c.Latitude>=-90&&c.Latitude<=90&&c.Longitude>=-180&&c.Longitude<=180}
func validatePolygon(p []models.Coordinate)string{if len(p)<4{return"polygon must contain at least 4 points"};if p[0]!=p[len(p)-1]{return"polygon must be closed (first point must equal last point)"};for _,p:=range p{if !validCoordinate(p){return"polygon contains an invalid latitude or longitude"}};return""}
func(a *API) Geofences(w http.ResponseWriter,r *http.Request){start:=time.Now();if r.Method=="GET"{items,err:=a.Store.ListGeofences(r.Context());if err!=nil{a.error(w,500,"could not list geofences",start);return};a.respond(w,200,items,start);return};var g models.Geofence;if err:=decode(r,&g);err!=nil{a.error(w,400,"invalid JSON body",start);return};if strings.TrimSpace(g.Name)==""{a.error(w,400,"name is required",start);return};if !models.Categories[g.Category]{a.error(w,400,"category must be restricted, commercial, residential, or industrial",start);return};if msg:=validatePolygon(g.Polygon);msg!=""{a.error(w,400,msg,start);return};created,err:=a.Store.CreateGeofence(r.Context(),g);if err!=nil{a.error(w,500,"could not create geofence",start);return};a.respond(w,201,created,start)}
func(a *API) Vehicles(w http.ResponseWriter,r *http.Request){start:=time.Now();if r.Method=="GET"{items,err:=a.Store.ListVehicles(r.Context());if err!=nil{a.error(w,500,"could not list vehicles",start);return};a.respond(w,200,items,start);return};var v models.Vehicle;if err:=decode(r,&v);err!=nil||strings.TrimSpace(v.Name)==""{a.error(w,400,"name is required",start);return};created,err:=a.Store.CreateVehicle(r.Context(),v);if err!=nil{a.error(w,500,"could not create vehicle",start);return};a.respond(w,201,created,start)}
func(a *API) Location(w http.ResponseWriter,r *http.Request){start:=time.Now();vehicleID:=strings.TrimPrefix(r.URL.Path,"/api/vehicles/");vehicleID=strings.TrimSuffix(vehicleID,"/location");var loc models.LocationUpdate;if err:=decode(r,&loc);err!=nil||loc.Latitude < -90||loc.Latitude > 90||loc.Longitude < -180||loc.Longitude > 180{a.error(w,400,"invalid latitude or longitude",start);return};if err:=a.Geo.ProcessLocation(r.Context(),vehicleID,loc);err!=nil{if errors.Is(err,service.ErrVehicleNotFound){a.error(w,404,"vehicle not found",start)}else{a.error(w,500,"could not process location",start)};return};a.respond(w,200,map[string]string{"status":"location recorded"},start)}
func(a *API) Alerts(w http.ResponseWriter,r *http.Request){start:=time.Now();q:=r.URL.Query();limit:=100;if v:=q.Get("limit");v!=""{if n,err:=strconv.Atoi(v);err==nil&&n>0{limit=n}};since:=time.Time{};if v:=q.Get("since");v!=""{if t,err:=time.Parse(time.RFC3339,v);err==nil{since=t}};items,err:=a.Store.ListAlertsFiltered(r.Context(),models.AlertFilter{VehicleID:q.Get("vehicle_id"),GeofenceID:q.Get("geofence_id"),EventType:q.Get("event_type"),Since:since},limit);if err!=nil{a.error(w,500,"could not list alerts",start);return};a.respond(w,200,items,start)}
func(a *API) Rules(w http.ResponseWriter,r *http.Request){start:=time.Now();var rule models.AlertRule;if err:=decode(r,&rule);err!=nil||rule.GeofenceID==""||(rule.EventType!="entry"&&rule.EventType!="exit"&&rule.EventType!="both"){a.error(w,400,"geofence_id and event_type (entry, exit, or both) are required",start);return};created,err:=a.Store.CreateRule(r.Context(),rule);if err!=nil{a.error(w,500,"could not create alert rule",start);return};a.respond(w,201,created,start)}
func(a *API) WebSocket(w http.ResponseWriter,r *http.Request){c,err:=a.Upgrader.Upgrade(w,r,nil);if err!=nil{return};a.Hub.Add(c);defer a.Hub.Remove(c);for{if _,_,err:=c.ReadMessage();err!=nil{return}}}
