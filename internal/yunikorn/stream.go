package yunikorn

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"

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
	// TODO: This is Okayish for small number of events, but for large number of events this will be a bottleneck
	// We should consider using a channel? or a pool of workers? or a different queuing system ? to handle events.
	if err := s.eventHandler(ctx, &eventRecord); err != nil {
		logger.Errorf("error handling event: %v", err)
	}

	if err := s.eventRepository.Record(ctx, &eventRecord); err != nil {
		logger.Errorf("error recording event: %v", err)
	}

	if eventRecord.GetType() == si.EventRecord_APP {
		printAppEvent(&eventRecord)
	}

	return nil
}

func printAppEvent(er *si.EventRecord) {
	fmt.Printf("---------\n")
	fmt.Printf("Type         : %s\n", si.EventRecord_Type_name[int32(er.GetType())])
	fmt.Printf("ObjectId     : %s\n", er.GetObjectID())
	fmt.Printf("Message      : %s\n", er.GetMessage())
	fmt.Printf("Change Type  : %s\n", er.GetEventChangeType())
	fmt.Printf("Change Detail: %s\n", er.GetEventChangeDetail())
	fmt.Printf("Reference ID:  %s\n", er.GetReferenceID())
	fmt.Printf("Resource    : %+v\n", er.GetResource())
}
