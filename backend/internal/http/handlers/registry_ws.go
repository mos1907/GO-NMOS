package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// wsConn is a thin wrapper so we don't expose the websocket type in Handler directly.
type wsConn struct {
	*websocket.Conn
}

type RegistryEvent struct {
	Kind      string      `json:"kind"`
	Resource  string      `json:"resource"`
	Info      interface{} `json:"info,omitempty"`
	Timestamp string      `json:"timestamp"`
}

var registryUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// CORS is already handled at HTTP layer; for WebSocket we allow same host.
		return true
	},
}

// RegistryEventsWS upgrades the connection to WebSocket and subscribes the client
// to NMOS registry events (best-effort).
func (h *Handler) RegistryEventsWS(w http.ResponseWriter, r *http.Request) {
	conn, err := registryUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	wc := &wsConn{Conn: conn}

	h.registryMu.Lock()
	h.registryConns[wc] = struct{}{}
	h.registryMu.Unlock()

	// Keep the connection alive until the client closes it.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	h.registryMu.Lock()
	delete(h.registryConns, wc)
	h.registryMu.Unlock()
	_ = conn.Close()
}

// publishRegistryEvent broadcasts a registry event to all connected websocket clients.
func (h *Handler) publishRegistryEvent(ev RegistryEvent) {
	if h == nil {
		return
	}

	h.registryMu.Lock()
	defer h.registryMu.Unlock()
	if len(h.registryConns) == 0 {
		return
	}

	if ev.Timestamp == "" {
		ev.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	data, err := json.Marshal(ev)
	if err != nil {
		return
	}

	for c := range h.registryConns {
		_ = c.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			_ = c.Close()
			delete(h.registryConns, c)
		}
	}
}
