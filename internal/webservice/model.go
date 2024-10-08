package webservice

import (
	"encoding/json"
	"fmt"
)

// ProblemDetails represents the problem details as per RFC 7807
type ProblemDetails struct {
	// Type is a URI reference that identifies the problem type.
	Type string `json:"type,omitempty"`
	// Title is a short, human-readable summary of the problem type.
	Title string `json:"title,omitempty"`
	// Status is the HTTP status code generated by the origin server for this occurrence of the problem.
	Status int `json:"status,omitempty"`
	// Detail is a human-readable explanation specific to this occurrence of the problem.
	Detail string `json:"detail,omitempty"`
	// Instance is a URI reference that identifies the specific occurrence of the problem.
	Instance string `json:"instance,omitempty"`
}

func (pd *ProblemDetails) Error() string {
	val, err := json.MarshalIndent(pd, "", "  ")
	if err != nil {
		return fmt.Sprintf("%s: %s", pd.Title, pd.Detail)
	}
	return string(val)
}
