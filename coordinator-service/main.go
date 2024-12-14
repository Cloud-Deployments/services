package main

import (
	"fmt"
	"github.com/Cloud-Deployments/services/coordinator/queue"
	"github.com/Cloud-Deployments/services/coordinator/runner"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {

	// create a new manager for every organization
	manager := runner.NewManager()

	go func() {
		manager.Run()
	}()

	handler := queue.NewHandler(manager, queue.HandlerOpts{
		Driver: queue.NewRedisDriver(queue.RedisDriverOpts{
			Host:     "cache",
			Port:     6379,
			Password: "",
			DB:       0,
			Prefix:   "coordinator-service_",
		}),
	})

	go handler.Run()

	http.HandleFunc("/ws/{organizationId}", manager.ListenForRunners)
	port := fmt.Sprintf(":%s", os.Getenv("WS_PORT"))
	log.Println("WebSocket server started on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
