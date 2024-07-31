package yunikorn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-history-server/internal/database/repository"

	"go.uber.org/mock/gomock"

	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
	"github.com/stretchr/testify/assert"
)

func TestFetchEventStream(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepository := repository.NewMockRepository(mockCtrl)
	eventRepository := repository.NewInMemoryEventRepository()
	mockYunikornClient := NewMockClient(mockCtrl)
	mockYunikornClient.EXPECT().GetEventStream(gomock.Any()).DoAndReturn(
		func(ctx context.Context) (*http.Response, error) {
			// Create a pipe to simulate the server streaming response
			reader, writer := io.Pipe()

			// Write events to the writer in a separate goroutine
			go func() {
				defer func() { _ = writer.Close() }()
				time.Sleep(50 * time.Millisecond) // Simulate streaming delay
				events := []*si.EventRecord{
					{Type: si.EventRecord_APP, EventChangeType: si.EventRecord_ADD},
					{Type: si.EventRecord_APP, EventChangeType: si.EventRecord_ADD},
					{Type: si.EventRecord_APP, EventChangeType: si.EventRecord_SET},
				}
				enc := json.NewEncoder(writer)
				for _, event := range events {
					if err := enc.Encode(event); err != nil {
						if err = writer.CloseWithError(err); err != nil {
							t.Errorf("error closing writer: %v", err)
						}
						return
					}
					time.Sleep(50 * time.Millisecond) // Simulate streaming delay
				}
			}()

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       reader,
			}, nil
		},
	)

	service := Service{
		repo:            mockRepository,
		eventRepository: eventRepository,
		client:          mockYunikornClient,
		eventHandler:    noopEventHandler,
	}

	// Start the ProcessEvents function in a separate goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	t.Cleanup(cancel)

	go func() {
		if err := service.ProcessEvents(ctx); err != nil {
			t.Errorf("error processing events: %v", err)
		}
	}()

	assert.Eventually(t, func() bool {
		eventCounts, err := service.eventRepository.Counts(ctx)
		if err != nil {
			t.Fatalf("error getting event counts: %v", err)
		}
		expectedKey1 := fmt.Sprintf("%s-%s", si.EventRecord_APP.String(), si.EventRecord_ADD.String())
		expectedKey2 := fmt.Sprintf("%s-%s", si.EventRecord_APP.String(), si.EventRecord_SET.String())
		return eventCounts[expectedKey1] == 2 && eventCounts[expectedKey2] == 1
	}, 1*time.Second, 50*time.Millisecond)
}

func TestProcessStreamResponse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedErr   error
		expectedType  si.EventRecord_Type
		expectedEvent si.EventRecord_ChangeType
		expectedCount int
	}{
		{
			name:          "Valid Event",
			input:         `{"type": 2, "eventChangeType": 2}` + "\n",
			expectedErr:   nil,
			expectedType:  si.EventRecord_APP,
			expectedEvent: si.EventRecord_ADD,
			expectedCount: 1,
		},
		{
			name:        "Invalid JSON",
			input:       `{"type": 2, "eventChangeType": 2` + "\n", // Invalid JSON (missing closing brace)
			expectedErr: errors.New("could not unmarshal event from stream"),
		},
		{
			name:          "Empty Input",
			input:         "",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				eventRepository: repository.NewInMemoryEventRepository(),
				eventHandler:    noopEventHandler,
			}

			err := service.processStreamResponse(context.Background(), []byte(tt.input))

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				if err != nil {
					t.Errorf("expected no error; got '%v'", err)
				}

				eventCounts, err := service.eventRepository.Counts(context.Background())
				if err != nil {
					t.Fatalf("error getting event counts: %v", err)
				}
				expectedKey := fmt.Sprintf("%s-%s", tt.expectedType.String(), tt.expectedEvent.String())
				assert.Equal(t, tt.expectedCount, eventCounts[expectedKey])
			}
		})
	}
}

func noopEventHandler(ctx context.Context, event *si.EventRecord) error {
	return nil
}
