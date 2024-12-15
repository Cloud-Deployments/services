package api

import (
	"encoding/json"
	"fmt"
	"github.com/Cloud-Deployments/services/coordinator/job"
	"github.com/Cloud-Deployments/services/coordinator/queue"
	"log"
	"net/http"
	"os"
)

type HttpAPIOpts struct {
	Host string
	Port int
}

type HttpAPI struct {
	HttpAPIOpts

	queueHandler *queue.Handler
}

func NewHttpAPIServer(queueHandler *queue.Handler, opts HttpAPIOpts) *HttpAPI {
	return &HttpAPI{
		HttpAPIOpts:  opts,
		queueHandler: queueHandler,
	}
}

func (h *HttpAPI) Run() {
	http.HandleFunc("/jobs", h.handleNewJob)
	addr := fmt.Sprintf("%s:%d", h.Host, h.Port)
	log.Printf("API server started on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (h *HttpAPI) handleNewJob(w http.ResponseWriter, r *http.Request) {
	apiSecret := r.Header.Get("X-Api-Token")
	if apiSecret == "" || apiSecret != os.Getenv("API_TOKEN") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var j job.Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, "Failed to decode job", http.StatusBadRequest)
		return
	}

	if err := h.queueHandler.Dispatch(&j); err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to enqueue job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
