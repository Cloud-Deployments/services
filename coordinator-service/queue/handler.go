package queue

import (
	"github.com/Cloud-Deployments/services/coordinator/job"
	"github.com/Cloud-Deployments/services/coordinator/runner"
	"log"
	"time"
)

type HandlerOpts struct {
	Driver Driver
}

type Handler struct {
	HandlerOpts
	manager *runner.Manager

	jobs    []*job.Job
	closech chan struct{}
}

func NewHandler(manager *runner.Manager, opts HandlerOpts) *Handler {
	return &Handler{
		HandlerOpts: opts,

		manager: manager,
		jobs:    make([]*job.Job, 0),
		closech: make(chan struct{}),
	}
}

func (h *Handler) Close() {
	h.Driver.Close()
	close(h.closech)
}

func (h *Handler) Run() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-h.closech:
			log.Println("gracefully shutting down queue handler")
			return
		case <-ticker.C:
			j, err := h.Driver.Dequeue()
			if err != nil {
				log.Println("no jobs to process")
				continue
			}

			log.Printf("processing job: %s", j.Id)
		}
	}
}
