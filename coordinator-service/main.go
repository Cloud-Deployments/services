package main

import (
	"fmt"
	"github.com/Cloud-Deployments/services/coordinator/api"
	"github.com/Cloud-Deployments/services/coordinator/queue"
	"github.com/Cloud-Deployments/services/coordinator/runner"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {

	// create a new manager for every organization
	manager := runner.NewManager()

	go func() {
		manager.Run()
	}()

	redisPort, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		log.Fatal("REDIS_PORT must be an integer")
	}

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal("REDIS_DB must be an integer")
	}

	handler := queue.NewHandler(manager, queue.HandlerOpts{
		Driver: queue.NewRedisDriver(queue.RedisDriverOpts{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     redisPort,
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
			Prefix:   os.Getenv("REDIS_PREFIX"),
		}),
	})

	go handler.Run()

	httpPort, err := strconv.Atoi(os.Getenv("API_PORT"))
	if err != nil {
		log.Fatal("API_PORT must be an integer")
	}
	apiServer := api.NewHttpAPIServer(handler, api.HttpAPIOpts{
		Host: os.Getenv("API_HOST"),
		Port: httpPort,
	})
	go apiServer.Run()

	http.HandleFunc("/ws/{organizationId}", manager.ListenForRunners)
	wsPort := fmt.Sprintf(":%s", os.Getenv("WS_PORT"))
	log.Println("WebSocket server started on", wsPort)
	log.Fatal(http.ListenAndServe(wsPort, nil))
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
