package realtime

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"mapup/backend/internal/models"
)

type Hub struct { mu sync.RWMutex; clients map[*websocket.Conn]struct{} }
func NewHub()*Hub{return &Hub{clients:make(map[*websocket.Conn]struct{})}}
func(h *Hub) Add(c *websocket.Conn){h.mu.Lock();h.clients[c]=struct{}{};h.mu.Unlock()}
func(h *Hub) Remove(c *websocket.Conn){h.mu.Lock();delete(h.clients,c);h.mu.Unlock();_ = c.Close()}
func(h *Hub) Broadcast(a models.Alert){payload,_:=json.Marshal(a);h.mu.RLock();clients:=make([]*websocket.Conn,0,len(h.clients));for c:=range h.clients{clients=append(clients,c)};h.mu.RUnlock();for _,c:=range clients{if err:=c.WriteMessage(websocket.TextMessage,payload);err!=nil{h.Remove(c)}}}
