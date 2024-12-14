package runner

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 4,
	WriteBufferSize: 1024 * 4,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Manager struct {
	pools     map[string]*Pool
	broadcast chan *Message
}

type Message struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

func NewManager() *Manager {
	return &Manager{
		pools:     map[string]*Pool{},
		broadcast: make(chan *Message),
	}
}

func (m *Manager) ListenForRunners(w http.ResponseWriter, r *http.Request) {
	organizationId := r.PathValue("organizationId")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}

	pool, ok := m.pools[organizationId]
	if !ok {
		pool = NewPool(organizationId, m)
		m.pools[organizationId] = pool
		go pool.Run()
	}

	// Read the runner's ID and token
	runner := &Runner{
		OrganizationId: organizationId,
		Addr:           conn.RemoteAddr(),
		Pool:           pool,
		Conn:           conn,
		Send:           make(chan []byte, 256),
		JoinedAt:       time.Now(),
		Authenticated:  false,
	}
	pool.register <- runner

	go runner.authLoop()
	go runner.readPump()
	go runner.writePump()
}

func (m *Manager) Run() {
	for {
		select {
		case message := <-m.broadcast:
			switch message.Type {
			case "ping":
				fmt.Println("Received ping message")
				continue
			}
		}
	}
}
