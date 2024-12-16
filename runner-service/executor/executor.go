package executor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Cloud-Deployments/services/runner/job"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
)

type ExecutorOpts struct {
}

type Executor struct {
	ExecutorOpts
}

func NewExecutor(opts ExecutorOpts) *Executor {
	return &Executor{
		ExecutorOpts: opts,
	}
}

func (e *Executor) Run(j *job.Job, conn *websocket.Conn) error {
	config := &ssh.ClientConfig{
		User: j.Connection.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(j.Connection.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         0,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", j.Connection.Host, j.Connection.Port), config)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			logLine := scanner.Text()
			e.sendLog(j, conn, "stdout", logLine, false)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logLine := scanner.Text()
			e.sendLog(j, conn, "stderr", logLine, false)
		}
	}()

	err = session.Start(j.Command)
	if err != nil {
		log.Fatalf("failed to start command: %s", err)
		return err
	}

	err = session.Wait()
	if err != nil {
		log.Fatalf("failed to wait for command: %s", err)
		return err
	}

	return e.sendLog(j, conn, "stdout", "Command execution finished", true)
}

type LogRequest struct {
	RunnerId string `json:"runner_id"`
	JobId    string `json:"job_id"`
	Finished bool   `json:"finished"`
	Type     string `json:"type"`
	Log      string `json:"log"`
}

func (e *Executor) sendLog(j *job.Job, conn *websocket.Conn, logType string, logLine string, finished bool) error {
	type req struct {
		Type string `json:"type"`
		Data []byte `json:"data"`
	}

	logRequest := LogRequest{
		RunnerId: os.Getenv("RUNNER_ID"),
		JobId:    j.Id,
		Finished: finished,
		Type:     logType,
		Log:      logLine,
	}

	logData, err := json.Marshal(logRequest)
	if err != nil {
		log.Println("Error marshalling log data:", err)
		return err
	}

	data, err := json.Marshal(&req{
		Type: "job-log",
		Data: logData,
	})
	if err != nil {
		log.Println("Error marshalling request data:", err)
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}
