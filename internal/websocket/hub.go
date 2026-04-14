package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"serverManage/internal/logger"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	id       string
	service  string
	socket   *websocket.Conn
	send     chan []byte
	hub      *Hub
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	log        *logger.Logger
	mu         sync.RWMutex
}

func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		log:        log,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastLog(serviceName, message string, level logger.LogLevel) {
	entry := logger.LogEntry{
		ServiceName: serviceName,
		Level:       level,
		Message:     message,
		Timestamp:   time.Now(),
	}

	data, _ := json.Marshal(entry)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.service == "" || client.service == serviceName {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, serviceName string) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		id:      r.RemoteAddr,
		service: serviceName,
		socket:  socket,
		send:    make(chan []byte, 256),
		hub:     h,
	}

	h.register <- client

	go client.write()
	go client.read()
}

func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			break
		}

		// 可以处理客户端消息，如订阅/取消订阅
		var msg map[string]string
		if json.Unmarshal(message, &msg) == nil {
			if action, ok := msg["action"]; ok {
				if action == "subscribe" {
					c.service = msg["service"]
				}
			}
		}
	}
}

func (c *Client) write() {
	defer c.socket.Close()

	for {
		message, ok := <-c.send
		if !ok {
			c.socket.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.socket.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}