package workers

import (
	"bufio"
	"bytes"
	"github.com/SarthakMakhija/blast-core/payload"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/SarthakMakhija/blast-core/report"
)

type BytesWriteCloser struct {
	*bufio.Writer
}

func (writeCloser *BytesWriteCloser) Close() error {
	return nil
}

func TestWritesPayloadByWorker(t *testing.T) {
	loadGenerationResponse := make(chan report.LoadGenerationResponse)

	var buffer bytes.Buffer
	worker := Worker{
		connection: &BytesWriteCloser{bufio.NewWriter(&buffer)},
		requestId:  NewRequestId(),
		options: WorkerOptions{
			maxDuration:            2 * time.Millisecond,
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
		},
	}

	go func() {
		for response := range loadGenerationResponse {
			assert.Nil(t, response.Err)
			assert.Equal(t, int64(7), response.PayloadLengthBytes)
		}
	}()

	var workerWaitGroup sync.WaitGroup
	workerWaitGroup.Add(1)

	worker.run(&workerWaitGroup)
	workerWaitGroup.Wait()

	close(loadGenerationResponse)
}

func TestWritesMultiplePayloadsByWorker(t *testing.T) {
	loadGenerationResponse := make(chan report.LoadGenerationResponse)

	var buffer bytes.Buffer
	worker := Worker{
		connection: &BytesWriteCloser{bufio.NewWriter(&buffer)},
		requestId:  NewRequestId(),
		options: WorkerOptions{
			maxDuration:            2 * time.Millisecond,
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
		},
	}

	go func() {
		for response := range loadGenerationResponse {
			assert.Nil(t, response.Err)
			assert.Equal(t, int64(7), response.PayloadLengthBytes)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	close(loadGenerationResponse)
}

func TestWritesMultiplePayloadsByWorkerWithThrottle(t *testing.T) {
	loadGenerationResponse := make(chan report.LoadGenerationResponse)

	var buffer bytes.Buffer
	worker := Worker{
		connection: &BytesWriteCloser{bufio.NewWriter(&buffer)},
		requestId:  NewRequestId(),
		options: WorkerOptions{
			maxDuration:            2 * time.Millisecond,
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
			requestsPerSecond:      float64(3),
		},
	}

	go func() {
		for response := range loadGenerationResponse {
			assert.Nil(t, response.Err)
			assert.Equal(t, int64(7), response.PayloadLengthBytes)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	close(loadGenerationResponse)
}

func TestWritesOnANilConnectionWithConnectionId(t *testing.T) {
	loadGenerationResponse := make(chan report.LoadGenerationResponse)

	worker := Worker{
		connection: nil,
		requestId:  NewRequestId(),
		options: WorkerOptions{
			maxDuration:            2 * time.Millisecond,
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
		},
	}

	go func() {
		for response := range loadGenerationResponse {
			assert.Error(t, response.Err)
			assert.Equal(t, -1, response.ConnectionId)
			assert.Equal(t, ErrNilConnection, response.Err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	close(loadGenerationResponse)
}

func TestWritesPayloadByWorkerWithConnectionId(t *testing.T) {
	loadGenerationResponse := make(chan report.LoadGenerationResponse)

	var buffer bytes.Buffer
	worker := Worker{
		connection:   &BytesWriteCloser{bufio.NewWriter(&buffer)},
		requestId:    NewRequestId(),
		connectionId: 10,
		options: WorkerOptions{
			maxDuration:            2 * time.Millisecond,
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
		},
	}

	go func() {
		for response := range loadGenerationResponse {
			assert.Nil(t, response.Err)
			assert.Equal(t, 10, response.ConnectionId)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	close(loadGenerationResponse)
}
