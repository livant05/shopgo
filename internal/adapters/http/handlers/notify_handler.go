package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// NotifyHub fans out byte messages to all subscribed SSE clients.
type NotifyHub struct {
	mu   sync.RWMutex
	subs map[string]chan []byte
}

func NewNotifyHub() *NotifyHub {
	return &NotifyHub{subs: make(map[string]chan []byte)}
}

func (h *NotifyHub) subscribe(id string) chan []byte {
	ch := make(chan []byte, 8)
	h.mu.Lock()
	h.subs[id] = ch
	h.mu.Unlock()
	return ch
}

func (h *NotifyHub) unsubscribe(id string) {
	h.mu.Lock()
	if ch, ok := h.subs[id]; ok {
		close(ch)
		delete(h.subs, id)
	}
	h.mu.Unlock()
}

// Broadcast sends data to every connected SSE client.
func (h *NotifyHub) Broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.subs {
		select {
		case ch <- data:
		default: // drop if slow client
		}
	}
}

// NotifyHandler serves the SSE stream endpoint.
type NotifyHandler struct{ hub *NotifyHub }

func NewNotifyHandler(hub *NotifyHub) *NotifyHandler { return &NotifyHandler{hub} }

func (h *NotifyHandler) Stream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	clientID := fmt.Sprintf("%p", c.Request)
	ch := h.hub.subscribe(clientID)
	defer h.hub.unsubscribe(clientID)

	// Send a keepalive comment so the browser confirms the connection
	fmt.Fprintf(c.Writer, ": connected\n\n")
	c.Writer.Flush()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
			c.Writer.(http.Flusher).Flush()
		}
	}
}
