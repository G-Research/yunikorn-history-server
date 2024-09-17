package webservice

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
)

const (
	queryParamSubmissionStartTime = "submissionStartTime"
	queryParamSubmissionEndTime   = "submissionEndTime"
	queryParamGroups              = "groups"
	queryParamLimit               = "limit"
	queryParamOffset              = "offset"
	queryParamUser                = "user"
)

func parseApplicationFilters(r *http.Request) (*repository.ApplicationFilters, error) {
	filters := repository.ApplicationFilters{}
	user := getUserQueryParam(r)
	if user != "" {
		filters.User = &user
	}
	submissionStartTime, err := getSubmissionStartTimeQueryParam(r)
	if err != nil {
		return nil, err
	}
	if submissionStartTime != nil {
		filters.SubmissionStartTime = submissionStartTime
	}
	submissionEndTime, err := getSubmissionEndTimeQueryParam(r)
	if err != nil {
		return nil, err
	}
	if submissionEndTime != nil {
		filters.SubmissionEndTime = submissionEndTime
	}
	offset, err := getOffsetQueryParam(r)
	if err != nil {
		return nil, err
	}
	if offset != nil {
		filters.Offset = offset
	}
	limit, err := getLimitQueryParam(r)
	if err != nil {
		return nil, err
	}
	if limit != nil {
		filters.Limit = limit
	}
	groups := getGroupsQueryParam(r)
	if len(groups) > 0 {
		filters.Groups = groups
	}
	return &filters, nil
}

func getUserQueryParam(r *http.Request) string {
	return r.URL.Query().Get(queryParamUser)
}

func getGroupsQueryParam(r *http.Request) []string {
	groups := r.URL.Query().Get(queryParamGroups)
	groupsSlice := strings.Split(groups, ",")
	return groupsSlice
}

func getOffsetQueryParam(r *http.Request) (*int, error) {
	offsetStr := r.URL.Query().Get(queryParamOffset)
	if offsetStr == "" {
		return nil, nil
	}

	return toInt(offsetStr)
}

func getLimitQueryParam(r *http.Request) (*int, error) {
	limitStr := r.URL.Query().Get(queryParamLimit)
	if limitStr == "" {
		return nil, nil
	}

	return toInt(limitStr)
}

func toInt(numberString string) (*int, error) {
	offset, err := strconv.Atoi(numberString)
	if err != nil {
		return nil, fmt.Errorf("invalid 'offset' query parameter: %v", err)
	}

	return &offset, nil
}

func getSubmissionStartTimeQueryParam(r *http.Request) (*time.Time, error) {
	startStr := r.URL.Query().Get(queryParamSubmissionStartTime)
	if startStr == "" {
		return nil, nil
	}

	return toTime(startStr)
}

func getSubmissionEndTimeQueryParam(r *http.Request) (*time.Time, error) {
	endStr := r.URL.Query().Get(queryParamSubmissionEndTime)
	if endStr == "" {
		return nil, nil
	}

	return toTime(endStr)
}

func toTime(millisString string) (*time.Time, error) {
	startMillis, err := strconv.ParseInt(millisString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid 'start' query parameter: %v", err)
	}

	// Convert milliseconds since epoch to a time.Time object
	startTime := time.UnixMilli(startMillis)
	return &startTime, nil
}
