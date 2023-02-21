package main

import (
	"fmt"
	"sync"
	"time"
)

type networkThresholds struct {
	*sync.Mutex
	thresholds map[int64]int64 // map of network ID to threshold in seconds
}

func main() {

	thresholds := networkThresholds{
		Mutex: &sync.Mutex{},
		thresholds: map[int64]int64{
			1: 1,
			2: 20,
		},
	}

	mockLastSeen := map[int64]int64{
		1: time.Now().Unix(),
		2: time.Now().Unix(),
	}

	go func() {
		squelcher(thresholds)
	}()

	for {
		for networkID, threshold := range thresholds.thresholds {
			sinceSeen := time.Now().Unix() - mockLastSeen[networkID]
			if sinceSeen > threshold {
				fmt.Printf("Network %d has not been seen in %d seconds\n",
					networkID, sinceSeen)
			}
		}
		time.Sleep(1 * time.Second)
	}
}
