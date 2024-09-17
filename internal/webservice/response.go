package webservice

import (
	"net/http"

	restful "github.com/emicklei/go-restful/v3"

	"github.com/G-Research/yunikorn-history-server/internal/log"
)

// jsonResponse writes the data to the response writer as a JSON object.
func jsonResponse(response *restful.Response, data any) {
	if err := response.WriteHeaderAndJson(http.StatusOK, data, "application/json"); err != nil {
		log.Logger.Errorf("could not write response: %v", err)
	}
}

// errorResponse writes an RFC7807 Problem error response to the response writer.
func errorResponse(req *restful.Request, resp *restful.Response, err error) {
	log.Logger.Errorf("error processing request for %s: %v", req.Request.URL.Path, err)
	problemDetails := ProblemDetails{
		Type:     "about:blank",
		Title:    "Internal Server Error",
		Status:   http.StatusInternalServerError,
		Detail:   err.Error(),
		Instance: req.Request.URL.Path,
	}
	if err := resp.WriteHeaderAndJson(http.StatusInternalServerError, problemDetails, "application/problem+json"); err != nil {
		log.Logger.Errorf("could not write error response: %v", err)
	}
}

func badRequestResponse(req *restful.Request, resp *restful.Response, err error) {
	log.Logger.Errorf("error processing request for %s: %v", req.Request.URL.Path, err)
	problemDetails := ProblemDetails{
		Type:     "about:blank",
		Title:    "Bad Request",
		Status:   http.StatusBadRequest,
		Detail:   err.Error(),
		Instance: req.Request.URL.Path,
	}
	if err := resp.WriteHeaderAndJson(http.StatusBadRequest, problemDetails, "application/problem+json"); err != nil {
		log.Logger.Errorf("could not write error response: %v", err)
	}
}
