package tests

import (
	"github.com/SarthakMakhija/blast-core/payload"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/SarthakMakhija/blast-core/report"
	"github.com/SarthakMakhija/blast-core/workers"
)

func TestSendsRequestsWithSingleConnection(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:8080", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency := uint(10)
	workerGroup := workers.NewWorkerGroup(workers.NewGroupOptions(concurrency, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:8080", 2*time.Millisecond))
	loadGenerationResponseChannel := workerGroup.Run()

	go func() {
		for response := range loadGenerationResponseChannel {
			assert.Nil(t, response.Err)
			assert.Equal(t, int64(10), response.PayloadLengthBytes)
		}
	}()

	workerGroup.WaitTillDone()
	close(loadGenerationResponseChannel)
}

func TestSendsRequestsWithMultipleConnections(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:8081", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, connections := uint(20), uint(10)
	workerGroup := workers.NewWorkerGroup(workers.NewGroupOptionsFullyLoaded(concurrency, connections, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:8081", 3*time.Second, 1, 2*time.Millisecond))
	loadGenerationResponseChannel := workerGroup.Run()
	uniqueConnectionIds := make(map[int]bool)

	go func() {
		for response := range loadGenerationResponseChannel {
			uniqueConnectionIds[response.ConnectionId] = true
			assert.Nil(t, response.Err)
			assert.Equal(t, int64(10), response.PayloadLengthBytes)
		}
	}()

	workerGroup.WaitTillDone()
	close(loadGenerationResponseChannel)

	connectionIds := make([]int, 0, len(uniqueConnectionIds))
	for connectionId := range uniqueConnectionIds {
		connectionIds = append(connectionIds, connectionId)
	}
	sort.Ints(connectionIds)

	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, connectionIds)
}

func TestSendsARequestAndReadsResponseWithSingleConnection(t *testing.T) {
	payloadSizeBytes, responseSizeBytes := int64(10), int64(10)
	server, err := NewEchoServer("tcp", "localhost:8082", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)

	concurrency, totalRequests := uint(10), uint(20)
	responseChannel := make(chan report.SubjectServerResponse)

	defer func() {
		server.stop()
		close(responseChannel)
	}()

	workerGroup := workers.NewWorkerGroupWithResponseReader(
		workers.NewGroupOptions(concurrency, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:8082", 2*time.Millisecond), report.NewResponseReader(responseSizeBytes, 100*time.Millisecond, responseChannel),
	)
	loadGenerationResponseChannel := workerGroup.Run()

	totalRequestsSent := 0
	go func() {
		for response := range loadGenerationResponseChannel {
			totalRequests = totalRequests + 1
			assert.Nil(t, response.Err)
			assert.Equal(t, int64(10), response.PayloadLengthBytes)
		}
	}()

	workerGroup.WaitTillDone()
	close(loadGenerationResponseChannel)

	for count := 1; count < totalRequestsSent; count++ {
		response := <-responseChannel
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(10), response.PayloadLengthBytes)
	}
}

func TestSendsRequestsOnANonRunningServer(t *testing.T) {
	concurrency := uint(10)

	workerGroup := workers.NewWorkerGroup(workers.NewGroupOptions(concurrency, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:8090", 2*time.Millisecond))
	loadGenerationResponseChannel := workerGroup.Run()

	go func() {
		for response := range loadGenerationResponseChannel {
			assert.Error(t, response.Err)
			assert.Equal(t, workers.ErrNilConnection, response.Err)
		}
	}()

	workerGroup.WaitTillDone()
	close(loadGenerationResponseChannel)
}

func TestSendsRequestsWithDialTimeout(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:8098", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency := uint(1)

	workerGroup := workers.NewWorkerGroup(workers.NewGroupOptionsFullyLoaded(concurrency, 1, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:8098", 1*time.Millisecond, 1.0, 2*time.Millisecond))
	loadGenerationResponseChannel := workerGroup.Run()

	go func() {
		for response := range loadGenerationResponseChannel {
			assert.Error(t, response.Err)
		}
	}()

	workerGroup.WaitTillDone()
	close(loadGenerationResponseChannel)
}
