package community

import (
	"sync"

	"chat/utils"
)

// client represents a live websocket connection subscribed to a channel.
type client struct {
	channelId int
	conn      *utils.WebSocket
}

// hub maintains live websocket clients grouped by channel id and broadcasts
// new messages to every subscriber of the channel.
type hub struct {
	mu      sync.RWMutex
	clients map[int]map[*client]struct{}
}

var defaultHub = &hub{clients: map[int]map[*client]struct{}{}}

func (h *hub) register(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.channelId] == nil {
		h.clients[c.channelId] = map[*client]struct{}{}
	}
	h.clients[c.channelId][c] = struct{}{}
}

func (h *hub) unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.clients[c.channelId]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clients, c.channelId)
		}
	}
}

// broadcast sends a message to every live client subscribed to the channel.
func (h *hub) broadcast(channelId int, payload wsOutgoing) {
	h.mu.RLock()
	set := h.clients[channelId]
	clients := make([]*client, 0, len(set))
	for c := range set {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		if c.conn != nil && !c.conn.IsClosed() {
			c.conn.Send(payload)
		}
	}
}
