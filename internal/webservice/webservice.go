package webservice

import (
	"context"
	"encoding/json"
	"net/http"
	"richscott/yhs/internal/config"
	"richscott/yhs/internal/repository"

	"github.com/julienschmidt/httprouter"
)

type WebService struct {
	server  *http.Server
	storage *repository.RepoPostgres
}

func NewWebService(addr string, storage *repository.RepoPostgres) *WebService {
	return &WebService{
		server: &http.Server{
			Addr: addr,
		},
		storage: storage,
	}
}

func (ws *WebService) Start(ctx context.Context) error {
	router := httprouter.New()
	router.Handle("GET", PARTITIONS, ws.getPartitions)
	router.Handle("GET", QUEUES_PER_PARTITION, ws.getQueuesPerPartition)
	router.Handle("GET", APPS_PER_PARTITION_PER_QUEUE, ws.getAppsPerPartitionPerQueue)
	router.Handle("GET", NODES_PER_PARTITION, ws.getNodesPerPartition)
	ws.server.Handler = router
	return ws.server.ListenAndServe()
}

func (ws *WebService) getPartitions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	partitions, err := ws.storage.GetAllPartitions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(&config.PartitionsResponse{Partitions: partitions})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ws *WebService) getQueuesPerPartition(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	partition := params.ByName("partition_name")
	queues, err := ws.storage.GetQueuesPerPartition(partition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(&config.QueuesResponse{Queues: queues})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ws *WebService) getAppsPerPartitionPerQueue(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	partition := params.ByName("partition_name")
	queue := params.ByName("queue_name")
	apps, err := ws.storage.GetAppsPerPartitionPerQueue(partition, queue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(&config.AppsResponse{Apps: apps})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ws *WebService) getNodesPerPartition(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	partition := params.ByName("partition_name")
	nodes, err := ws.storage.GetNodesPerPartition(partition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(nodes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
