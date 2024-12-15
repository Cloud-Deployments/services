package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Cloud-Deployments/services/coordinator/job"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Runner struct {
	OrganizationId string
	Id             string
	Token          string
	Addr           net.Addr
	Pool           *Pool
	Conn           *websocket.Conn
	Send           chan []byte
	JoinedAt       time.Time

	Authenticated bool
	Available     bool
}

func (r *Runner) authLoop() {
	var (
		interval = 1 * time.Second
		attempts = 5
	)

	for {
		if r.Authenticated {
			log.Println("Runner authenticated", r.Id)
			break
		}

		if attempts == 0 {
			log.Println("Runner failed to authenticate", r.Id)
			r.Conn.Close()
			r.Pool.unregister <- nil
			break
		}

		attempts--
		time.Sleep(interval)
	}
}

func (r *Runner) readPump() {
	defer func() {
		r.Pool.unregister <- nil
		r.Conn.Close()
	}()
	r.Conn.SetReadLimit(maxMessageSize)
	r.Conn.SetReadDeadline(time.Now().Add(pongWait))
	r.Conn.SetPongHandler(func(string) error { r.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := r.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		var msg *Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		if msg.Type == "auth" {
			err = r.Authenticate(msg.Data)
			if err != nil {
				log.Println("Error authenticating runner:", err)
			}

			continue
		}

		r.Pool.Manager.broadcast <- msg
	}
}

func (r *Runner) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		r.Conn.Close()
	}()

	r.Available = true
	for {
		select {
		case message, ok := <-r.Send:
			r.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				r.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := r.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(r.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-r.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			r.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := r.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type NewJobMessage struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

func (r *Runner) SendJob(j *job.Job) error {
	jobData, err := json.Marshal(&j)
	if err != nil {
		log.Println("Error marshalling job data:", err)
		return err
	}

	newJobMsg := NewJobMessage{
		Type: "new-job",
		Data: jobData,
	}

	jsonData, err := json.Marshal(newJobMsg)
	if err != nil {
		log.Println("Error marshalling new job message:", err)
		return err
	}

	r.Available = false
	r.Send <- jsonData

	return nil
}

type AuthCredentialsRequest struct {
	RunnerId    string `json:"runner_id"`
	RunnerToken string `json:"runner_token"`
}

func (r *Runner) Authenticate(data []byte) error {
	var authReq AuthCredentialsRequest
	err := json.Unmarshal(data, &authReq)
	if err != nil {
		return err
	}

	if authReq.RunnerId == "" || authReq.RunnerToken == "" {
		return errors.New("invalid credentials")
	}

	// check creds with api and check if has access to organization
	r.Id = authReq.RunnerId
	r.Token = authReq.RunnerToken
	r.Authenticated = true

	return nil
}
