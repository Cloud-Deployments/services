package connection

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Cloud-Deployments/services/runner/executor"
	"github.com/Cloud-Deployments/services/runner/job"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

type CoordinatorConnection struct {
	CoordinatorConnectionOpts

	quitch chan struct{}
	client *websocket.Conn
}

type AuthCredentials struct {
	RunnerId    string `json:"runner_id"`
	RunnerToken string `json:"runner_token"`
}

type CoordinatorConnectionOpts struct {
	Executor *executor.Executor

	Host string
	Port int
	Path string

	AuthCredentials AuthCredentials
}

type Message struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

func NewCoordinatorConnection(opts CoordinatorConnectionOpts) *CoordinatorConnection {
	return &CoordinatorConnection{
		CoordinatorConnectionOpts: opts,
		quitch:                    make(chan struct{}),
	}
}

func (c *CoordinatorConnection) Close() {
	close(c.quitch)
}

func isBase64Encoded(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func (c *CoordinatorConnection) Connect() error {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.Path,
	}

	client, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer client.Close()

	c.client = client

	go func() {
		err := c.Authenticate()
		if err != nil {
			log.Println("auth", err)
			return
		}
	}()

	go func() {
		defer c.Close()
		for {
			_, message, err := client.ReadMessage()
			if err != nil {
				log.Println("read", err)
				continue
			}

			var msg Message
			err = json.Unmarshal(message, &msg)
			if err != nil {
				log.Println("unmarshal", err)
				continue
			}

			switch msg.Type {
			case "new-job":
				log.Println("Received new job message")

				//if isBase64Encoded(string(msg.Data)) {
				//	msg.Data, err = base64.StdEncoding.DecodeString(string(msg.Data))
				//	if err != nil {
				//		log.Println("Error decoding base64 data:", err)
				//		continue
				//	}
				//}

				var j job.Job
				err := json.Unmarshal(msg.Data, &j)
				if err != nil {
					log.Println("Error unmarshalling job:", err)
					continue
				}

				fmt.Printf("job data recv: %+v\n", j)

				go func() {
					err := c.Executor.Run(&j, client)
					if err != nil {
						log.Println("Error running job:", err)
					}
				}()
			}

			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.quitch:
			log.Println("gracefully closing connection")
			err := client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return err
			}

			return nil
		case <-ticker.C:
			err := client.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("write", err)
				return err
			}
		}
	}
}

func (c *CoordinatorConnection) Authenticate() error {
	credData, err := json.Marshal(c.AuthCredentials)
	if err != nil {
		return err
	}

	authRequestData := Message{
		Type: "auth",
		Data: credData,
	}

	jsonData, err := json.Marshal(authRequestData)
	if err != nil {
		return err
	}

	return c.client.WriteMessage(websocket.TextMessage, jsonData)
}
