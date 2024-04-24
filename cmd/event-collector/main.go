package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	//"github.com/apache/yunikorn-scheduler-interface/lib/go/common"
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
	// "github.com/sanity-io/litter"
)

func main() {
	httpProto := "http"
	ykHost := "127.0.0.1"
	ykPort := 9889
	streamEndPt := "/ws/v1/events/stream"

	streamURL := fmt.Sprintf("%s://%s:%d%s", httpProto, ykHost, ykPort, streamEndPt)

	resp, err := http.Get(streamURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not request from %s: %v", streamURL, err)
		os.Exit(1)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not read from http stream: %v", err)
			break
		}

		ev := si.EventRecord{}
		err = json.Unmarshal(line, &ev)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not unmarshal event from stream: %v\n", err)
			break
		}

		fmt.Printf("---------\n")
		fmt.Printf("Type         : %s\n", si.EventRecord_Type_name[int32(ev.Type)])
		fmt.Printf("ObjectId     : %s\n", ev.ObjectID)
		fmt.Printf("Message      : %s\n", ev.Message)
		fmt.Printf("Change Type  : %s\n", ev.EventChangeType)
		fmt.Printf("Change Detail: %s\n", ev.EventChangeDetail)
		fmt.Printf("Reference ID:  %s\n", ev.ReferenceID)
		// if ev.Resource != nil {
		// 	fmt.Printf("Resource:      %s\n", litter.Sdump(*ev.Resource))
		// }
	}
}
