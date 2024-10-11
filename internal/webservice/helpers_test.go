package webservice

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/util"
)

func TestGetUserQueryParam(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		result string
	}{
		{"No user param", "", ""},
		{"With user param", "user=john", "john"},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", "/?"+tt.query, nil)
		if err != nil {
			t.Fatal(err)
		}
		result := getUserQueryParam(req)
		if result != tt.result {
			t.Errorf("expected %v, got %v", tt.result, result)
		}
	}
}

func TestGetGroupsQueryParam(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		result []string
	}{
		{"No groups param", "", nil},
		{"Single group", "groups=admin", []string{"admin"}},
		{"Multiple groups", "groups=admin,user,guest", []string{"admin", "user", "guest"}},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", "/?"+tt.query, nil)
		require.NoError(t, err)

		result := getGroupsQueryParam(req)
		require.Equal(t, tt.result, result)
	}
}

func TestGetOffsetQueryParam(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		result *int
		hasErr bool
	}{
		{"No offset param", "", nil, false},
		{"Valid offset", "offset=5", util.ToPtr(5), false},
		{"Invalid offset", "offset=abc", nil, true},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", "/?"+tt.query, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, err := getOffsetQueryParam(req)
		if (err != nil) != tt.hasErr {
			t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
		}
		if result != nil && *result != *tt.result {
			t.Errorf("expected %v, got %v", tt.result, result)
		}
	}
}

func TestGetLimitQueryParam(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		result *int
		hasErr bool
	}{
		{"No limit param", "", nil, false},
		{"Valid limit", "limit=10", util.ToPtr(10), false},
		{"Invalid limit", "limit=xyz", nil, true},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", "/?"+tt.query, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, err := getLimitQueryParam(req)
		if (err != nil) != tt.hasErr {
			t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
		}
		if result != nil && *result != *tt.result {
			t.Errorf("expected %v, got %v", tt.result, result)
		}
	}
}

func TestGetStartQueryParam(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		result *time.Time
		hasErr bool
	}{
		{"No start param", "", nil, false},
		{"Valid start", "submissionStartTime=1625097600000", util.ToPtr(time.UnixMilli(1625097600000)), false},
		{"Invalid start", "submissionStartTime=invalid", nil, true},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", "/?"+tt.query, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, err := getSubmissionStartTimeQueryParam(req)
		if (err != nil) != tt.hasErr {
			t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
		}
		if result != nil && !result.Equal(*tt.result) {
			t.Errorf("expected %v, got %v", tt.result, result)
		}
	}
}

func TestGetEndQueryParam(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		result *time.Time
		hasErr bool
	}{
		{"No end param", "", nil, false},
		{"Valid end", "submissionEndTime=1625097600000", util.ToPtr(time.UnixMilli(1625097600000)), false},
		{"Invalid end", "submissionEndTime=invalid", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := http.NewRequest("GET", "/?"+tt.query, nil)
			if err != nil {
				t.Fatal(err)
			}
			result, err := getSubmissionEndTimeQueryParam(req)
			if (err != nil) != tt.hasErr {
				t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
			}
			if result != nil && !result.Equal(*tt.result) {
				t.Errorf("expected %v, got %v", tt.result, result)
			}
		})
	}
}

func TestGetTimestampStartQueryParam(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		result *time.Time
		hasErr bool
	}{
		{"No start param", "", nil, false},
		{"Valid start", "timestampStart=1625097600000", util.ToPtr(time.UnixMilli(1625097600000)), false},
		{"Invalid start", "timestampStart=invalid", nil, true},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", "/?"+tt.query, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, err := getTimestampStartQueryParam(req)
		if (err != nil) != tt.hasErr {
			t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
		}
		if result != nil && !result.Equal(*tt.result) {
			t.Errorf("expected %v, got %v", tt.result, result)
		}
	}
}

func TestGetTimestampEndQueryParam(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		result *time.Time
		hasErr bool
	}{
		{"No start param", "", nil, false},
		{"Valid start", "timestampEnd=1625097600000", util.ToPtr(time.UnixMilli(1625097600000)), false},
		{"Invalid start", "timestampEnd=invalid", nil, true},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", "/?"+tt.query, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, err := getTimestampEndQueryParam(req)
		if (err != nil) != tt.hasErr {
			t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
		}
		if result != nil && !result.Equal(*tt.result) {
			t.Errorf("expected %v, got %v", tt.result, result)
		}
	}
}
