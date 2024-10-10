package yunikorn

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/G-Research/yunikorn-scheduler-interface/lib/go/si"

	"github.com/G-Research/yunikorn-history-server/internal/log"
)

func (s *Service) ProcessEvents(ctx context.Context) error {
	logger := log.FromContext(ctx)

	resp, err := s.client.GetEventStream(ctx)
	if err != nil {
		return fmt.Errorf("error getting event stream: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Errorf("error closing event stream response body: %v", err)
		}
	}()

	reader := bufio.NewReader(resp.Body)
	for {
		response, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := s.processStreamResponse(ctx, response); err != nil {
			return fmt.Errorf("error processing stream response: %w", err)
		}
	}
}

func (s *Service) processStreamResponse(ctx context.Context, response []byte) error {
	logger := log.FromContext(ctx)

	if len(response) == 0 {
		logger.Warn("empty response from yunikorn event stream")
		return nil
	}

	var eventRecord si.EventRecord
	if err := json.Unmarshal(response, &eventRecord); err != nil {
		return fmt.Errorf("could not unmarshal event from stream: %w", err)
	}

	logger.Infow(
		"received event from yunikorn event stream",
		"type", si.EventRecord_Type_name[int32(eventRecord.GetType())],
		"objectId", eventRecord.GetObjectID(),
		"message", eventRecord.GetMessage(),
		"change_type", eventRecord.GetEventChangeType(),
		"change_detail", eventRecord.GetEventChangeDetail(),
		"reference_id", eventRecord.GetReferenceID(),
		"resource", eventRecord.GetResource(),
		"state", eventRecord.GetState(),
	)

	// TODO: This is Okayish for small number of events, but for large number of events this will be a bottleneck
	// We should consider using a channel? or a pool of workers? or a different queuing system ? to handle events.
	if err := s.eventHandler(ctx, &eventRecord); err != nil {
		logger.Errorf("error handling event: %v", err)
	}

	if err := s.eventRepository.Record(ctx, &eventRecord); err != nil {
		logger.Errorf("error recording event: %v", err)
	}

	return nil
}
