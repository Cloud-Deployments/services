package main

import (
	"fmt"
	"github.com/Cloud-Deployments/services/runner/connection"
	"github.com/Cloud-Deployments/services/runner/executor"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func main() {
	port, err := strconv.Atoi(os.Getenv("COORDINATOR_PORT"))
	if err != nil {
		log.Fatal("COORDINATOR_PORT must be an integer")
	}

	coordinator := connection.NewCoordinatorConnection(connection.CoordinatorConnectionOpts{
		Executor: executor.NewExecutor(executor.ExecutorOpts{
			//
		}),
		Host: os.Getenv("COORDINATOR_HOST"),
		Port: port,
		Path: fmt.Sprintf("/ws/%s", os.Getenv("ORGANIZATION_ID")),
		AuthCredentials: connection.AuthCredentials{
			RunnerId:    os.Getenv("RUNNER_ID"),
			RunnerToken: os.Getenv("RUNNER_TOKEN"),
		},
	})

	log.Fatal(coordinator.Connect())
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
