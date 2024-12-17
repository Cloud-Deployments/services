package main

import (
	"fmt"
	"github.com/Cloud-Deployments/services/runner/connection"
	"github.com/Cloud-Deployments/services/runner/executor"
	"log"
	"os"
)

type Config struct {
	CoordinatorHost string
	CoordinatorPort int
}

func main() {
	config := Config{
		CoordinatorHost: "127.0.0.1",
		CoordinatorPort: 4001,
	}

	runnerId := os.Getenv("RUNNER_ID")
	runnerToken := os.Getenv("RUNNER_TOKEN")
	organizationId := os.Getenv("ORGANIZATION_ID")

	if runnerId == "" || runnerToken == "" || organizationId == "" {
		log.Fatal("RUNNER_ID, RUNNER_TOKEN and ORGANIZATION_ID must be set")
	}

	coordinator := connection.NewCoordinatorConnection(connection.CoordinatorConnectionOpts{
		Executor: executor.NewExecutor(executor.ExecutorOpts{
			//
		}),
		Host: config.CoordinatorHost,
		Port: config.CoordinatorPort,
		Path: fmt.Sprintf("/ws/%s", organizationId),
		AuthCredentials: connection.AuthCredentials{
			RunnerId:    runnerId,
			RunnerToken: runnerToken,
		},
	})

	log.Fatal(coordinator.Connect())
}

//func init() {
//	err := godotenv.Load()
//	if err != nil {
//		log.Fatal("Error loading .env file")
//	}
//}
