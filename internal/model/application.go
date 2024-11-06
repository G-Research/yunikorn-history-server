package model

import (
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

type Application struct {
	Metadata               `json:",inline"`
	dao.ApplicationDAOInfo `json:",inline"`
}

func (app *Application) MergeFrom(appInfo *dao.ApplicationDAOInfo) {
	app.ID = appInfo.ID
	app.PartitionID = appInfo.PartitionID
	app.Partition = appInfo.Partition
	app.QueueID = appInfo.QueueID
	app.QueueName = appInfo.QueueName
	app.SubmissionTime = appInfo.SubmissionTime
	app.FinishedTime = appInfo.FinishedTime
	app.State = appInfo.State
	app.User = appInfo.User
	app.Groups = appInfo.Groups
	app.RejectedMessage = appInfo.RejectedMessage
	app.StateLog = appInfo.StateLog
	app.HasReserved = appInfo.HasReserved
	app.Reservations = appInfo.Reservations
	app.MaxRequestPriority = appInfo.MaxRequestPriority
	app.StartTime = appInfo.StartTime
	app.PlaceholderData = appInfo.PlaceholderData

	lookup := make(map[string]struct{})
	if len(appInfo.Allocations) > 0 {
		for _, alloc := range app.Allocations {
			lookup[alloc.AllocationKey] = struct{}{}
		}
		for _, alloc := range appInfo.Allocations {
			if _, ok := lookup[alloc.AllocationKey]; !ok {
				app.Allocations = append(app.Allocations, alloc)
			}
		}
		clear(lookup)
	}
	if len(appInfo.Requests) > 0 {
		for _, ask := range app.Requests {
			lookup[ask.AllocationKey] = struct{}{}
		}
		for _, ask := range appInfo.Requests {
			if _, ok := lookup[ask.AllocationKey]; !ok {
				app.Requests = append(app.Requests, ask)
			}
		}
		clear(lookup)
	}
	if appInfo.UsedResource != nil {
		app.UsedResource = appInfo.UsedResource
	}
	if appInfo.MaxUsedResource != nil {
		app.MaxUsedResource = appInfo.MaxUsedResource
	}
	if appInfo.PendingResource != nil {
		app.PendingResource = appInfo.PendingResource
	}
	if appInfo.ResourceUsage != nil {
		app.ResourceUsage = appInfo.ResourceUsage
	}
	if appInfo.PreemptedResource != nil {
		app.PreemptedResource = appInfo.PreemptedResource
	}
	if appInfo.PlaceholderResource != nil {
		app.PlaceholderResource = appInfo.PlaceholderResource
	}
}
