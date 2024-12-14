package main

import (
	"github.com/Cloud-Deployments/services/runner/connection"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	coordinator := connection.NewCoordinatorConnection(connection.CoordinatorConnectionOpts{
		Host: "coordinator-service",
		Port: 4001,
		Path: "/ws/123",
		AuthCredentials: connection.AuthCredentials{
			RunnerId:    "runner-1",
			RunnerToken: "runner-1-token",
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
