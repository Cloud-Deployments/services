package executor

import (
	"fmt"
	"github.com/Cloud-Deployments/services/runner/job"
	"golang.org/x/crypto/ssh"
	"log"
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

func (e *Executor) Run(j *job.Job) ([]byte, error) {
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
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, j.Command)
	}()

	output, err := session.CombinedOutput("/bin/sh")
	if err != nil {
		log.Fatalf("failed to execute command: %s", err)
		return nil, err
	}

	fmt.Println("job response", string(output))
	return output, nil
}
