package update

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var (
	lock chan bool
)

// New is a factory function creating a new  Handler instance
func New(updateFn func(map[string]string), updateLock chan bool) *Handler {
	if updateLock != nil {
		lock = updateLock
	} else {
		lock = make(chan bool, 1)
		lock <- true
	}

	return &Handler{
		fn:   updateFn,
		Path: "/v1/update",
	}
}

// Handler is an API handler used for triggering container update scans
type Handler struct {
	fn   func(containerImageTags map[string]string)
	Path string
}

type UpdateRequestBody struct {
	AiDriverTag string
	DaemonTag   string
}

// Handle is the actual http.Handle function doing all the heavy lifting
func (handle *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Info("Updates triggered by HTTP API request.")

	var urb UpdateRequestBody
	err := json.NewDecoder(r.Body).Decode(&urb)
	if err != nil {
		fmt.Fprintln(w, "Failed to load request body")
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("Update Skipped. %s", err)
		return
	}

	containerImageTags := make(map[string]string)
	containerImageTags["aidriver"] = urb.AiDriverTag
	containerImageTags["daemon"] = urb.DaemonTag

	select {
	case chanValue := <-lock:
		defer func() { lock <- chanValue }()
		handle.fn(containerImageTags)
	default:
		log.Debug("Skipped. Another update already running.")
	}

}
