package webservice

import (
	"encoding/json"
	"net/http"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"

	"github.com/G-Research/yunikorn-history-server/internal/model"

	"github.com/G-Research/yunikorn-history-server/internal/log"
)

type PartitionsResponse struct {
	Partitions []*dao.PartitionInfo `json:"partitions"`
}

type QueuesResponse struct {
	Queues []*model.PartitionQueueDAOInfo `json:"queues"`
}

type AppsResponse struct {
	Apps []*model.ApplicationDAOInfo `json:"apps"`
}

type NodesResponse struct {
	Nodes []*dao.NodeDAOInfo `json:"nodes"`
}

// jsonResponse writes the data to the response writer as a JSON object.
func jsonResponse(w http.ResponseWriter, data any) {
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Logger.Errorf("could not write response: %v", err)
	}
}

// errorResponse writes an RFC7807 Problem error response to the response writer.
func errorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Logger.Errorf("error processing request for %s: %v", r.URL.Path, err)
	problemDetails := ProblemDetails{
		Type:     "about:blank",
		Title:    "Internal Server Error",
		Status:   http.StatusInternalServerError,
		Detail:   err.Error(),
		Instance: r.URL.Path,
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(problemDetails); err != nil {
		log.Logger.Errorf("could not write error response: %v", err)
	}
}

func badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Logger.Errorf("error processing request for %s: %v", r.URL.Path, err)
	problemDetails := ProblemDetails{
		Type:     "about:blank",
		Title:    "Bad Request",
		Status:   http.StatusBadRequest,
		Detail:   err.Error(),
		Instance: r.URL.Path,
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(problemDetails); err != nil {
		log.Logger.Errorf("could not write error response: %v", err)
	}
}
