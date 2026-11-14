package main

import (
	"context"
	"log"
	"net/http"

	"mapup/backend/internal/config"
	"mapup/backend/internal/db"
	"mapup/backend/internal/handlers"
	"mapup/backend/internal/realtime"
	"mapup/backend/internal/service"
)

func main(){cfg:=config.Load();ctx:=context.Background();store,err:=db.New(ctx,cfg.DatabaseURL);if err!=nil{log.Fatal(err)};defer store.Pool.Close();hub:=realtime.NewHub();geo:=service.New(store,hub);api:=handlers.New(store,geo,hub);mux:=http.NewServeMux();mux.HandleFunc("/api/geofences",api.Geofences);mux.HandleFunc("/api/vehicles",api.Vehicles);mux.HandleFunc("/api/vehicles/",api.Location);mux.HandleFunc("/api/alerts",api.Alerts);mux.HandleFunc("/api/alert-rules",api.Rules);mux.HandleFunc("/ws",api.WebSocket);handler:=http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Header().Set("Access-Control-Allow-Origin",cfg.AllowedOrigin);w.Header().Set("Access-Control-Allow-Headers","Content-Type");w.Header().Set("Access-Control-Allow-Methods","GET,POST,OPTIONS");if r.Method=="OPTIONS"{w.WriteHeader(http.StatusNoContent);return};mux.ServeHTTP(w,r)});log.Printf("API listening on :%s",cfg.Port);log.Fatal(http.ListenAndServe(":"+cfg.Port,handler))}
