package log_buffer

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/seaweedfs/seaweedfs/weed/pb/filer_pb"
)

func TestNewLogBufferFirstBuffer(t *testing.T) {
	flushInterval := time.Second
	lb := NewLogBuffer("test", flushInterval, func(startTime, stopTime time.Time, buf []byte) {
		fmt.Printf("flush from %v to %v %d bytes\n", startTime, stopTime, len(buf))
	}, func() {
	})

	startTime := time.Now()

	messageSize := 1024
	messageCount := 5000

	receivedMessageCount := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		lastProcessedTime, isDone, err := lb.LoopProcessLogData("test", startTime, 0, func() bool {
			// stop if no more messages
			return receivedMessageCount < messageCount
		}, func(logEntry *filer_pb.LogEntry) error {
			receivedMessageCount++
			if receivedMessageCount >= messageCount {
				println("processed all messages")
				return io.EOF
			}
			return nil
		})

		fmt.Printf("before flush: sent %d received %d\n", messageCount, receivedMessageCount)
		fmt.Printf("lastProcessedTime %v isDone %v err: %v\n", lastProcessedTime, isDone, err)
		if err != nil && err != io.EOF {
			t.Errorf("unexpected error %v", err)
		}
	}()

	var buf = make([]byte, messageSize)
	for i := 0; i < messageCount; i++ {
		rand.Read(buf)
		lb.AddToBuffer(nil, buf, 0)
	}
	wg.Wait()

	if receivedMessageCount != messageCount {
		t.Errorf("expect %d messages, but got %d", messageCount, receivedMessageCount)
	}
}
