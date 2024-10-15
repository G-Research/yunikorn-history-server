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
	queryParamSubmissionStartTime          = "submissionStartTime"
	queryParamSubmissionEndTime            = "submissionEndTime"
	queryParamGroups                       = "groups"
	queryParamLimit                        = "limit"
	queryParamOffset                       = "offset"
	queryParamUser                         = "user"
	queryParamTimestampStart               = "timestampStart"
	queryParamTimestampEnd                 = "timestampEnd"
	queryParamPartition                    = "partition"
	queryParamNodeId                       = "nodeId"
	queryParamHostName                     = "hostName"
	queryParamRackName                     = "rackName"
	queryParamSchedulable                  = "schedulable"
	queryParamIsReserved                   = "isReserved"
	queryParamClusterID                    = "clusterId"
	queryParamState                        = "state"
	queryParamName                         = "name"
	queryParamLastStateTransitionTimeStart = "lastStateTransitionTimeStart"
	queryParamLastStateTransitionTimeEnd   = "lastStateTransitionTimeEnd"
)

func parsePartitionFilters(r *http.Request) (*repository.PartitionFilters, error) {
	var filters repository.PartitionFilters

	filters.Name = getNameQueryParam(r)
	filters.ClusterID = getClusterIDQueryParam(r)
	filters.State = getStateQueryParam(r)

	lastStateTransitionTimeStart, err := getLastStateTransitionTimeStartQueryParam(r)
	if err != nil {
		return nil, err
	}
	filters.LastStateTransitionTimeStart = lastStateTransitionTimeStart

	lastStateTransitionTimeEnd, err := getLastStateTransitionTimeEndQueryParam(r)
	if err != nil {
		return nil, err
	}
	filters.LastStateTransitionTimeEnd = lastStateTransitionTimeEnd

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

	return &filters, nil
}

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

func parseNodeUtilizationFilters(r *http.Request) (*repository.NodeUtilFilters, error) {
	var filters repository.NodeUtilFilters

	filters.ClusterID = getClusterIDQueryParam(r)
	filters.Partition = getPartitionQueryParam(r)
	limit, err := getLimitQueryParam(r)
	if err != nil {
		return nil, err
	}
	if limit != nil {
		filters.Limit = limit
	}
	offset, err := getOffsetQueryParam(r)
	if err != nil {
		return nil, err
	}
	if offset != nil {
		filters.Offset = offset
	}
	return &filters, nil
}

func parseHistoryFilters(r *http.Request) (*repository.HistoryFilters, error) {
	var filters repository.HistoryFilters
	timestampStart, err := getTimestampStartQueryParam(r)
	if err != nil {
		return nil, err
	}
	if timestampStart != nil {
		filters.TimestampStart = timestampStart
	}
	timestampEnd, err := getTimestampEndQueryParam(r)
	if err != nil {
		return nil, err
	}
	if timestampEnd != nil {
		filters.TimestampEnd = timestampEnd
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
	return &filters, nil
}

func parseNodeFilters(r *http.Request) (*repository.NodeFilters, error) {
	var filters repository.NodeFilters
	nodeId := getNodeIdQueryParam(r)
	if nodeId != "" {
		filters.NodeId = &nodeId
	}
	hostName := getHostNameQueryParam(r)
	if hostName != "" {
		filters.HostName = &hostName
	}
	rackName := getRackNameQueryParam(r)
	if rackName != "" {
		filters.RackName = &rackName
	}
	schedulable := getSchedulableQueryParam(r)
	if schedulable != nil {
		filters.Schedulable = schedulable
	}
	isReserved := getIsReservedQueryParam(r)
	if isReserved != nil {
		filters.IsReserved = isReserved
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
	return &filters, nil
}

func getNodeIdQueryParam(r *http.Request) string {
	return r.URL.Query().Get(queryParamNodeId)
}

func getHostNameQueryParam(r *http.Request) string {
	return r.URL.Query().Get(queryParamHostName)
}

func getRackNameQueryParam(r *http.Request) string {
	return r.URL.Query().Get(queryParamRackName)
}

func getSchedulableQueryParam(r *http.Request) *bool {
	schedulableStr := r.URL.Query().Get(queryParamSchedulable)
	schedulable, err := strconv.ParseBool(schedulableStr)
	if err != nil {
		return nil
	}
	return &schedulable
}

func getIsReservedQueryParam(r *http.Request) *bool {
	isReservedStr := r.URL.Query().Get(queryParamIsReserved)
	isReserved, err := strconv.ParseBool(isReservedStr)
	if err != nil {
		return nil
	}
	return &isReserved
}

func getClusterIDQueryParam(r *http.Request) *string {
	clusterId := r.URL.Query().Get(queryParamClusterID)
	if clusterId != "" {
		return &clusterId

	}
	return nil
}

func getPartitionQueryParam(r *http.Request) *string {
	partitionName := r.URL.Query().Get(queryParamPartition)
	if partitionName != "" {
		return &partitionName
	}
	return nil
}

func getStateQueryParam(r *http.Request) *string {
	state := r.URL.Query().Get(queryParamState)
	if state != "" {
		return &state

	}
	return nil
}
func getNameQueryParam(r *http.Request) *string {
	name := r.URL.Query().Get(queryParamName)
	if name != "" {
		return &name
	}
	return nil
}

func getUserQueryParam(r *http.Request) string {
	return r.URL.Query().Get(queryParamUser)
}

func getGroupsQueryParam(r *http.Request) []string {
	var groupsSlice []string
	groups := r.URL.Query().Get(queryParamGroups)
	if groups != "" {
		groupsSlice = strings.Split(groups, ",")
	}
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

func getLastStateTransitionTimeStartQueryParam(r *http.Request) (*time.Time, error) {
	startStr := r.URL.Query().Get(queryParamLastStateTransitionTimeStart)
	if startStr == "" {
		return nil, nil
	}

	return toTime(startStr)
}

func getLastStateTransitionTimeEndQueryParam(r *http.Request) (*time.Time, error) {
	endStr := r.URL.Query().Get(queryParamLastStateTransitionTimeEnd)
	if endStr == "" {
		return nil, nil
	}

	return toTime(endStr)
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

func getTimestampStartQueryParam(r *http.Request) (*time.Time, error) {
	startStr := r.URL.Query().Get(queryParamTimestampStart)
	if startStr == "" {
		return nil, nil
	}

	return toTime(startStr)
}

func getTimestampEndQueryParam(r *http.Request) (*time.Time, error) {
	endStr := r.URL.Query().Get(queryParamTimestampEnd)
	if endStr == "" {
		return nil, nil
	}

	return toTime(endStr)
}

func toTime(millisString string) (*time.Time, error) {
	startMillis, err := strconv.ParseInt(millisString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid 'timestamp' in query parameter: %v", err)
	}

	// Convert milliseconds since epoch to a time.Time object
	startTime := time.UnixMilli(startMillis)
	return &startTime, nil
}
