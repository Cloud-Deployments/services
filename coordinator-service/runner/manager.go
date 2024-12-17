package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Cloud-Deployments/services/coordinator/client"
	"github.com/Cloud-Deployments/services/coordinator/job"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 8,
	WriteBufferSize: 1024 * 8,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Manager struct {
	ApiClient *client.Client
	pools     map[string]*Pool
	broadcast chan *Message
}

type Message struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

func NewManager(apiClient *client.Client) *Manager {
	return &Manager{
		ApiClient: apiClient,
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
		Available:      false,
	}
	pool.register <- runner

	go runner.authLoop()
	go runner.readPump()
	go runner.writePump()
}

type JobLog struct {
	RunnerId string `json:"runner_id"`
	JobId    string `json:"job_id"`
	Finished bool   `json:"finished"`
	Type     string `json:"type"`
	Log      string `json:"log"`
}

func (m *Manager) Run() {
	for {
		select {
		case message := <-m.broadcast:
			switch message.Type {
			case "ping":
				fmt.Println("Received ping message")
				continue
			case "job-log":
				fmt.Println("Received job log message")
				var jobLog *JobLog
				err := json.Unmarshal(message.Data, &jobLog)
				if err != nil {
					log.Println("Error unmarshalling job response:", err)
					continue
				}

				go sendLogToWebhook(jobLog)

				if jobLog.Finished {
					for _, pool := range m.pools {
						for runner := range pool.runners {
							if runner.Id == jobLog.RunnerId {
								runner.Available = true
								break
							}
						}
					}
				}
			}
		}
	}
}

func (m *Manager) NewJob(j *job.Job) (bool, error) {
	for _, pool := range m.pools {
		for runner := range pool.runners {
			if runner.Available {
				err := runner.SendJob(j)
				if err != nil {
					continue
				}

				return true, nil
			}
		}
	}

	return false, nil
}

func sendLogToWebhook(jobLog *JobLog) {
	secret := os.Getenv("WEBHOOK_SECRET")
	url := os.Getenv("WEBHOOK_URL")

	jsonData, err := json.Marshal(jobLog)
	if err != nil {
		log.Println("Error marshalling job log data:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating webhook request:", err)
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", secret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending webhook request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		jsonBody, err := json.Marshal(resp.Body)
		if err != nil {
			log.Println("Error reading webhook response body:", err)
			return
		}

		log.Println("Webhook request failed:", resp.Status, string(jsonBody))
		return
	}
}
