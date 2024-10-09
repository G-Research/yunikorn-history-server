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

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/google/uuid"

	"github.com/G-Research/yunikorn-history-server/internal/health"
	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/G-Research/yunikorn-history-server/internal/model"
	ykmodel "github.com/G-Research/yunikorn-history-server/internal/yunikorn/model"
)

const (
	// routes
	routeClusters                 = "/api/v1/clusters"
	routePartitions               = "/api/v1/partitions"
	routeQueuesPerPartition       = "/api/v1/partition/{partition_name}/queues"
	routeAppsPerPartitionPerQueue = "/api/v1/partition/{partition_name}/queue/{queue_name}/applications"
	routeAppsHistory              = "/api/v1/history/apps"
	routeContainersHistory        = "/api/v1/history/containers"
	routeNodesPerPartition        = "/api/v1/partition/{partition_name}/nodes"
	routeNodeUtilization          = "/api/v1/scheduler/node-utilizations"
	routeSchedulerHealthcheck     = "/api/v1/scheduler/healthcheck"
	routeEventStatistics          = "/api/v1/event-statistics"
	routeHealthLiveness           = "/api/v1/health/liveness"
	routeHealthReadiness          = "/api/v1/health/readiness"

	// params
	paramsPartitionName = "partition_name"
	paramsQueueName     = "queue_name"
)

var startupTime = time.Now()

func (ws *WebService) init(ctx context.Context) {
	service := new(restful.WebService)

	service.Route(
		service.GET(routePartitions).
			To(ws.getPartitions).
			Produces(restful.MIME_JSON).
			Doc("Get all partitions").
			Writes([]dao.PartitionInfo{}).
			Returns(200, "OK", []dao.PartitionInfo{}).
			Returns(500, "Internal Server Error", ProblemDetails{}),
	)
	service.Route(
		service.GET(routeQueuesPerPartition).
			To(ws.getQueuesPerPartition).
			Param(
				service.PathParameter(
					"partition_name",
					"partition name",
				).
					DataType("string"),
			).
			Produces(restful.MIME_JSON).
			Writes([]*model.PartitionQueueDAOInfo{}).
			Returns(200, "OK", []*model.PartitionQueueDAOInfo{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get all queues for a partition"),
	)
	service.Route(
		service.GET(routeAppsPerPartitionPerQueue).
			To(ws.getAppsPerPartitionPerQueue).
			Param(
				service.PathParameter(
					"partition_name",
					"partition name",
				).
					DataType("string"),
			).
			Param(
				service.PathParameter(
					"queue_name",
					"queue name",
				).
					DataType("string"),
			).
			Produces(restful.MIME_JSON).
			Writes([]dao.ApplicationDAOInfo{}).
			Returns(200, "OK", []dao.ApplicationDAOInfo{}).
			Returns(400, "Bad Request", ProblemDetails{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get all applications for a partition and queue"),
	)
	service.Route(
		service.GET(routeNodesPerPartition).
			To(ws.getNodesPerPartition).
			Produces(restful.MIME_JSON).
			Writes([]dao.NodeDAOInfo{}).
			Returns(200, "OK", []dao.NodeDAOInfo{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get all nodes for a partition"),
	)
	service.Route(
		service.GET(routeAppsHistory).
			To(ws.getAppsHistory).
			Produces(restful.MIME_JSON).
			Writes([]dao.ApplicationHistoryDAOInfo{}).
			Returns(200, "OK", []dao.ApplicationHistoryDAOInfo{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get applications history"),
	)
	service.Route(
		service.GET(routeContainersHistory).
			To(ws.getContainersHistory).
			Produces(restful.MIME_JSON).
			Writes([]dao.ContainerHistoryDAOInfo{}).
			Returns(200, "OK", []dao.ContainerHistoryDAOInfo{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get containers history"),
	)
	service.Route(
		service.GET(routeNodeUtilization).
			To(ws.getNodeUtilizations).
			Produces(restful.MIME_JSON).
			Writes([]dao.PartitionNodesUtilDAOInfo{}).
			Returns(200, "OK", []dao.PartitionNodesUtilDAOInfo{}).
			Returns(500, "Internal Server Error", ProblemDetails{}).
			Doc("Get node utilization"),
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
			Title:       "Yunikorn History Server",
			Description: "Yunikorn History Server API",
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
	partitions, err := ws.repository.GetAllPartitions(ctx)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	jsonResponse(resp, partitions)
}

func (ws *WebService) getQueuesPerPartition(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	partition := req.PathParameter(paramsPartitionName)
	queues, err := ws.repository.GetQueuesPerPartition(ctx, partition)
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

func buildPartitionQueueTrees(ctx context.Context, queues []*model.PartitionQueueDAOInfo) ([]*model.PartitionQueueDAOInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(queues) == 0 {
		return nil, nil
	}

	queueMap := make(map[string]*model.PartitionQueueDAOInfo)
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

	roots := make([]*model.PartitionQueueDAOInfo, 0, len(rootIDs))
	for _, id := range rootIDs {
		roots = append(roots, queueMap[id])
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return roots, nil
}

// getAppsPerPartitionPerQueue returns all applications for a given partition and queue.
// Results are ordered by submission time in descending order.
// Following query params are supported:
// - user: filter by user
// - groups: filter by groups (comma-separated list)
// - submissionStartTime: filter from the submission time
// - submissionEndTime: filter until the submission time
// - limit: limit the number of returned applications
// - offset: offset the returned applications
func (ws *WebService) getAppsPerPartitionPerQueue(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	partition := req.PathParameter(paramsPartitionName)
	queue := req.PathParameter(paramsQueueName)

	filters, err := parseApplicationFilters(req.Request)
	if err != nil {
		badRequestResponse(req, resp, err)
		return
	}

	apps, err := ws.repository.GetAppsPerPartitionPerQueue(ctx, partition, queue, *filters)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	daoApps := make([]*dao.ApplicationDAOInfo, 0, len(apps))
	for _, app := range apps {
		daoApps = append(daoApps, &app.ApplicationDAOInfo)
	}

	jsonResponse(resp, daoApps)
}

func (ws *WebService) getNodesPerPartition(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	partition := req.PathParameter(paramsPartitionName)
	nodes, err := ws.repository.GetNodesPerPartition(ctx, partition)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	jsonResponse(resp, nodes)
}

// getAppsHistory returns history of Applications.
// Results are ordered by timestamp time in descending order.
// Following query params are supported:
// - timestampStart: filter from the timestamp
// - timestampEnd: filter until the timestamp
// - limit: limit the number of returned object
// - offset: offset the returned object
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

// getContainersHistory returns history of Containers.
// Results are ordered by timestamp time in descending order.
// Following query params are supported:
// - timestampStart: filter from the timestamp
// - timestampEnd: filter until the timestamp
// - limit: limit the number of returned object
// - offset: offset the returned object
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

func (ws *WebService) getNodeUtilizations(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	nodeUtilization, err := ws.repository.GetNodeUtilizations(ctx)
	if err != nil {
		errorResponse(req, resp, err)
		return
	}
	jsonResponse(resp, nodeUtilization)
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
