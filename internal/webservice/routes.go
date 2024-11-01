package webservice

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/google/uuid"

	"github.com/G-Research/unicorn-history-server/internal/health"
	"github.com/G-Research/unicorn-history-server/internal/log"
	"github.com/G-Research/unicorn-history-server/internal/model"
	ykmodel "github.com/G-Research/unicorn-history-server/internal/yunikorn/model"
)

const (
	// routes
	routeClusters                 = "/api/v1/clusters"
	routePartitions               = "/api/v1/partitions"
	routeQueuesPerPartition       = "/api/v1/partition/{partition_id}/queues"
	routeAppsPerPartitionPerQueue = "/api/v1/partition/{partition_id}/queue/{queue_id}/applications"
	routeAppsHistory              = "/api/v1/history/apps"
	routeContainersHistory        = "/api/v1/history/containers"
	routeNodesPerPartition        = "/api/v1/partition/{partition_id}/nodes"
	routeSchedulerHealthcheck     = "/api/v1/scheduler/healthcheck"
	routeEventStatistics          = "/api/v1/event-statistics"
	routeHealthLiveness           = "/api/v1/health/liveness"
	routeHealthReadiness          = "/api/v1/health/readiness"
)

var startupTime = time.Now()

func (ws *WebService) init(ctx context.Context) {
	service := new(restful.WebService)

	service.Route(
		service.GET(routePartitions).
			To(ws.getPartitions).
			Produces(restful.MIME_JSON).
			Doc("Get all partitions").
			Writes([]model.Partition{}).
			Param(service.QueryParameter("name", "Filter by partition name").DataType("string")).
			Param(service.QueryParameter("clusterId", "Filter by clusterId").DataType("string")).
			Param(service.QueryParameter("state", "Filter by state").DataType("string")).
			Param(service.QueryParameter(
				"lastStateTransitionTimeStart",
				"Filter from the lastStateTransitionTime (unix nanoseconds)",
			).DataType("string")).
			Param(service.QueryParameter(
				"lastStateTransitionTimeEnd",
				"Filter until the lastStateTransitionTime (unix nanoseconds)",
			).DataType("string")).
			Param(service.QueryParameter("limit", "Limit the number of returned partitions").DataType("int")).
			Param(service.QueryParameter("offset", "Offset the returned partitions").DataType("int")).
			Returns(200, "OK", []dao.PartitionInfo{}).
			Returns(500, "Internal Server Error", ProblemDetails{}),
	)
	service.Route(
		service.GET(routeQueuesPerPartition).
			To(ws.getQueuesPerPartition).
			Param(
				service.PathParameter(
					"partition_id",
					"partition id",
				).
					DataType("string"),
			).
			Produces(restful.MIME_JSON).
			Writes([]*model.Queue{}).
			Returns(200, "OK", []*model.Queue{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get all queues for a partition"),
	)
	service.Route(
		service.GET(routeAppsPerPartitionPerQueue).
			To(ws.getAppsPerPartitionPerQueue).
			Param(
				service.PathParameter(
					"partition_id",
					"partition id",
				).
					DataType("string"),
			).
			Param(
				service.PathParameter(
					"queue_id",
					"queue id",
				).
					DataType("string"),
			).
			Produces(restful.MIME_JSON).
			Writes([]dao.ApplicationDAOInfo{}).
			Param(service.QueryParameter("user", "Filter by user").DataType("string")).
			Param(service.QueryParameter("groups", "Filter by groups (comma-separated list)").
				DataType("string")).
			Param(service.QueryParameter("submissionStartTime", "Filter from the submission time (unix nanoseconds)").
				DataType("string")).
			Param(service.QueryParameter("submissionEndTime", "Filter until the submission time (unix nanoseconds)").
				DataType("string")).
			Param(service.QueryParameter("limit", "Limit the number of returned applications").DataType("int")).
			Param(service.QueryParameter("offset", "Offset the returned applications").DataType("int")).
			Returns(200, "OK", []dao.ApplicationDAOInfo{}).
			Returns(400, "Bad Request", ProblemDetails{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get all applications for a partition and queue"),
	)
	service.Route(
		service.GET(routeNodesPerPartition).
			To(ws.getNodesPerPartition).
			Param(
				service.PathParameter(
					"partition_id",
					"partition id",
				).
					DataType("string"),
			).
			Produces(restful.MIME_JSON).
			Writes([]model.Node{}).
			Param(service.QueryParameter("nodeId", "Filter by nodeId").DataType("string")).
			Param(service.QueryParameter("hostName", "Filter by hostName").DataType("string")).
			Param(service.QueryParameter("rackName", "Filter by rackName").DataType("string")).
			Param(service.QueryParameter("schedulable", "Filter by schedulable status").DataType("boolean")).
			Param(service.QueryParameter("isReserved", "Filter by reservation status").DataType("boolean")).
			Param(service.QueryParameter("limit", "Limit the number of returned nodes").DataType("int")).
			Param(service.QueryParameter("offset", "Offset the returned nodes").DataType("int")).
			Returns(200, "OK", []model.Node{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get all nodes for a partition"),
	)
	service.Route(
		service.GET(routeAppsHistory).
			To(ws.getAppsHistory).
			Produces(restful.MIME_JSON).
			Writes([]model.AppHistory{}).
			Param(service.QueryParameter("timestampStart", "Filter from the timestamp").DataType("string")).
			Param(service.QueryParameter("timestampEnd", "Filter until the timestamp").DataType("string")).
			Param(service.QueryParameter("limit", "Limit the number of returned objects").DataType("int")).
			Param(service.QueryParameter("offset", "Offset the returned objects").DataType("int")).
			Returns(200, "OK", []model.AppHistory{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get applications history"),
	)
	service.Route(
		service.GET(routeContainersHistory).
			To(ws.getContainersHistory).
			Produces(restful.MIME_JSON).
			Writes([]model.ContainerHistory{}).
			Param(service.QueryParameter("timestampStart", "Filter from the timestamp").DataType("string")).
			Param(service.QueryParameter("timestampEnd", "Filter until the timestamp").DataType("string")).
			Param(service.QueryParameter("limit", "Limit the number of returned objects").DataType("int")).
			Param(service.QueryParameter("offset", "Offset the returned objects").DataType("int")).
			Returns(200, "OK", []model.ContainerHistory{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get containers history"),
	)
	service.Route(
		service.GET(routeEventStatistics).
			To(ws.getEventStatistics).
			Produces(restful.MIME_JSON).
			Writes(ykmodel.EventTypeCounts{}).
			Returns(200, "OK", ykmodel.EventTypeCounts{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get event statistics"),
	)
	service.Route(
		service.GET(routeSchedulerHealthcheck).
			To(ws.LivenessHealthcheck).
			Produces(restful.MIME_JSON).
			Writes(health.LivenessStatus{}).
			Returns(200, "OK", health.LivenessStatus{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Scheduler liveness healthcheck"),
	)
	service.Route(
		service.GET(routeHealthLiveness).
			To(ws.LivenessHealthcheck).
			Produces(restful.MIME_JSON).
			Writes(health.LivenessStatus{}).
			Returns(200, "OK", health.LivenessStatus{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Liveness healthcheck"),
	)
	service.Route(
		service.GET(routeHealthReadiness).
			To(ws.ReadinessHealthcheck).
			Produces(restful.MIME_JSON).
			Writes(health.ReadinessStatus{}).
			Returns(200, "OK", health.ReadinessStatus{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Readiness healthcheck"),
	)

	container := restful.NewContainer()
	container.Add(service)

	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		APIPath:                       "/api/v1/openapi.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}
	container.Add(restfulspec.NewOpenAPIService(config))

	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.serveSPA)
	mux.HandleFunc("/swagger-ui/", ws.serveSwaggerUI)
	mux.HandleFunc("/api/v1/", container.ServeHTTP)

	ws.server.Handler = enrichRequestContextMiddleware(ctx, ws.applyCORS(mux))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Unicorn History Server",
			Description: "Unicorn History Server API",
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache 2.0",
					URL:  "http://www.apache.org/licenses/LICENSE-2.0.html",
				},
			},
			Version: "1.0.0",
		},
	}
}

func enrichRequestContextMiddleware(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.FromContext(ctx)
		rid := uuid.New().String()
		logger = logger.With("request_id", rid)
		ctx = log.ToContext(ctx, logger)
		*r = *r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *WebService) applyCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", strings.Join(s.config.CORSConfig.AllowedOrigins, ","))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(s.config.CORSConfig.AllowedMethods, ","))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(s.config.CORSConfig.AllowedHeaders, ","))

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ws *WebService) getPartitions(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	filters, err := parsePartitionFilters(req.Request)
	if err != nil {
		badRequestResponse(req, resp, err)
		return
	}
	partitions, err := ws.repository.GetAllPartitions(ctx, *filters)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	jsonResponse(resp, partitions)
}

func (ws *WebService) getQueuesPerPartition(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	partitionID := req.PathParameter("partition_id")
	queues, err := ws.repository.GetQueuesInPartition(ctx, partitionID)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	root, err := buildPartitionQueueTrees(ctx, queues)
	if err != nil {
		errorResponse(req, resp, err)
	}
	jsonResponse(resp, root)
}

func buildPartitionQueueTrees(ctx context.Context, queues []*model.Queue) ([]*model.Queue, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(queues) == 0 {
		return nil, nil
	}

	queueMap := make(map[string]*model.Queue)
	for _, queue := range queues {
		queueMap[queue.ID] = queue
	}

	var rootIDs []string
	for _, queue := range queues {
		if queue.ParentID == nil {
			rootIDs = append(rootIDs, queue.ID)
			continue
		}

		parent, ok := queueMap[*queue.ParentID]
		if !ok {
			return nil, fmt.Errorf("parent queue %q not found", queue.Parent)
		}
		parent.Children = append(parent.Children, queue)
	}

	if len(rootIDs) == 0 {
		return nil, fmt.Errorf("root queue not found")
	}

	sort.Strings(rootIDs)

	roots := make([]*model.Queue, 0, len(rootIDs))
	for _, id := range rootIDs {
		roots = append(roots, queueMap[id])
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return roots, nil
}

func (ws *WebService) getAppsPerPartitionPerQueue(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	partitionID := req.PathParameter("partition_id")
	queueID := req.PathParameter("queue_id")

	filters, err := parseApplicationFilters(req.Request)
	if err != nil {
		badRequestResponse(req, resp, err)
		return
	}

	apps, err := ws.repository.GetAppsPerPartitionPerQueue(ctx, partitionID, queueID, *filters)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}

	jsonResponse(resp, apps)
}

func (ws *WebService) getNodesPerPartition(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	partition := req.PathParameter("partition_id")
	filters, err := parseNodeFilters(req.Request)
	if err != nil {
		badRequestResponse(req, resp, err)
		return
	}
	nodes, err := ws.repository.GetNodesPerPartition(ctx, partition, *filters)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	jsonResponse(resp, nodes)
}

func (ws *WebService) getAppsHistory(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	filters, err := parseHistoryFilters(req.Request)
	if err != nil {
		badRequestResponse(req, resp, err)
		return
	}

	appsHistory, err := ws.repository.GetApplicationsHistory(ctx, *filters)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	jsonResponse(resp, appsHistory)
}

func (ws *WebService) getContainersHistory(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	filters, err := parseHistoryFilters(req.Request)
	if err != nil {
		badRequestResponse(req, resp, err)
		return
	}
	containersHistory, err := ws.repository.GetContainersHistory(ctx, *filters)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	jsonResponse(resp, containersHistory)
}

func (ws *WebService) getEventStatistics(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	counts, err := ws.eventRepository.Counts(ctx)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	jsonResponse(resp, counts)
}

func (ws *WebService) LivenessHealthcheck(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	jsonResponse(resp, ws.healthService.Liveness(ctx))
}

func (ws *WebService) ReadinessHealthcheck(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	jsonResponse(resp, ws.healthService.Readiness(ctx))
}

func (ws *WebService) serveSPA(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(ws.config.AssetsDir, r.URL.Path)
	fi, err := os.Stat(path)

	if os.IsNotExist(err) || fi.IsDir() {
		http.ServeFile(w, r, filepath.Join(ws.config.AssetsDir, "index.html"))
		return
	}

	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.FileServer(http.Dir(ws.config.AssetsDir)).ServeHTTP(w, r)
}

func (ws *WebService) serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	buf := bytes.NewReader([]byte(SwaggerUIHTML))
	http.ServeContent(w, r, "index.html", startupTime, buf)
}

const SwaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="description" content="SwaggerUI" />
    <title>SwaggerUI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
  </head>
	<body style="margin: 0">
	  <div id="swagger-ui"></div>
	  <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
	  <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js" crossorigin></script>
	  <script>
		window.onload = () => {
		  window.ui = SwaggerUIBundle({
			url: '/api/v1/openapi.json',
			dom_id: '#swagger-ui',
			presets: [
			  SwaggerUIBundle.presets.apis,
			  SwaggerUIStandalonePreset
			],
			layout: "StandaloneLayout",
		  });
		};
	  </script>
  </body>
</html>`
