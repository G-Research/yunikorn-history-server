package webservice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/G-Research/yunikorn-history-server/internal/config"
	"github.com/G-Research/yunikorn-history-server/internal/repository"
	"net/http"
	"os"
	"time"

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

func (ws *WebService) Start(ctx context.Context) {
	router := httprouter.New()
	router.Handle("GET", PARTITIONS, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r = r.WithContext(ctx)
		ws.getPartitions(w, r, p)
	})
	router.Handle("GET", QUEUES_PER_PARTITION, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r = r.WithContext(ctx)
		ws.getQueuesPerPartition(w, r, p)
	})
	router.Handle("GET", APPS_PER_PARTITION_PER_QUEUE, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r = r.WithContext(ctx)
		ws.getAppsPerPartitionPerQueue(w, r, p)
	})
	router.Handle("GET", NODES_PER_PARTITION, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r = r.WithContext(ctx)
		ws.getNodesPerPartition(w, r, p)
	})
	router.Handle("GET", APPS_HISTORY, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		r = r.WithContext(ctx)
		ws.getAppsHistory(w, r)
	})
	router.Handle("GET", CONTAINERS_HISTORY, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		r = r.WithContext(ctx)
		ws.getContainersHistory(w, r)
	})
	router.Handle("GET", NODE_UTILIZATION, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		r = r.WithContext(ctx)
		ws.getNodeUtilizations(w, r)
	})
	router.Handle("GET", EVENT_STATISTICS, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		r = r.WithContext(ctx)
		ws.getEventStatistics(w, r)
	})
	ws.server.Handler = router
	go func() {
		fmt.Printf("Starting webservice on %s\n", ws.server.Addr)
		err := ws.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP serving error: %v\n", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		fmt.Println("Shutting down webservice...")
		err := ws.server.Shutdown(shutdownCtx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "HTTP server shutdown error: %v\n", err)
		}
	}()
}

func (ws *WebService) getPartitions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	partitions, err := ws.storage.GetAllPartitions(r.Context())
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
	queues, err := ws.storage.GetQueuesPerPartition(r.Context(), partition)
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
	apps, err := ws.storage.GetAppsPerPartitionPerQueue(r.Context(), partition, queue)
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
	nodes, err := ws.storage.GetNodesPerPartition(r.Context(), partition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(nodes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ws *WebService) getAppsHistory(w http.ResponseWriter, r *http.Request) {
	appsHistory, err := ws.storage.GetApplicationsHistory(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(appsHistory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (ws *WebService) getContainersHistory(w http.ResponseWriter, r *http.Request) {
	containersHistory, err := ws.storage.GetContainersHistory(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(containersHistory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ws *WebService) getNodeUtilizations(w http.ResponseWriter, r *http.Request) {
	nodeUtilization, err := ws.storage.GetNodeUtilizations(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(nodeUtilization)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ws *WebService) getEventStatistics(w http.ResponseWriter, r *http.Request) {
	evCounts := r.Context().Value(config.EventCounts).(config.EventTypeCounts)
	if evCounts == nil {
		fmt.Fprintf(os.Stderr, "getEventStatistics(): could not get eventCounts map from context\n")
		return
	}

	evCountsStr := map[string]int{}
	for k, v := range evCounts {
		evCountsStr[k.String()] = v
	}

	err := json.NewEncoder(w).Encode(&evCountsStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
