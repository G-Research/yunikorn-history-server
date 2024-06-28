package info

import "time"

var (
	Version   string = "0.0.0-dev"
	BuildTime string = time.Now().Format(time.RFC3339)
	Commit    string = "unknown"
)
