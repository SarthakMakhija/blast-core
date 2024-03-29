package workers

import (
	"fmt"
	"github.com/SarthakMakhija/blast-core/report"
	"net"
	"os"
	"sync"
)

// WorkerGroup is a collection of workers that sends requestsPerRun to the server.
// WorkerGroup creates a total of GroupOptions.concurrency Workers.
// Each Worker sends WorkerOptions.requestsPerSecond requests per second.
// WorkerGroup also provides support for triggering response reading from the connection.
type WorkerGroup struct {
	options        GroupOptions
	stopChannel    chan struct{}
	doneChannel    chan struct{}
	responseReader *report.ResponseReader
	requestId      *RequestId
}

// NewWorkerGroup returns a new instance of WorkerGroup without supporting reading from the
// connection.
func NewWorkerGroup(options GroupOptions) *WorkerGroup {
	return NewWorkerGroupWithResponseReader(options, nil)
}

// NewWorkerGroupWithResponseReader returns a new instance of WorkerGroup
// that also supports reading from the connection.
func NewWorkerGroupWithResponseReader(
	options GroupOptions,
	responseReader *report.ResponseReader,
) *WorkerGroup {
	return &WorkerGroup{
		options:        options,
		stopChannel:    make(chan struct{}, options.concurrency),
		doneChannel:    make(chan struct{}, 1),
		responseReader: responseReader,
		requestId:      NewRequestId(),
	}
}

// Run runs the WorkerGroup and returns a channel of type report.LoadGenerationResponse.
// report.LoadGenerationResponse will contain each request sent by the Worker.
// This method runs a separate goroutine that runs the workers and the goroutine waits until
// all the workers are done.
func (group *WorkerGroup) Run() chan report.LoadGenerationResponse {
	loadGenerationResponseChannel := make(
		chan report.LoadGenerationResponse,
		group.options.ExpectedLoadInTotalDuration(),
	)

	go func() {
		group.runWorkers(loadGenerationResponseChannel)
		group.WaitTillDone()
		return
	}()
	return loadGenerationResponseChannel
}

// Close closes sends a stop signal to all the workers.
func (group *WorkerGroup) Close() {
	for count := 1; count <= int(group.options.concurrency); count++ {
		group.stopChannel <- struct{}{}
	}
}

// runWorkers runs all the workers.
// The numbers of workers that will run is determined by the concurrency field in GroupOptions.
// These workers will share the tcp connections and the sharing of tcp connections is determined
// by the number of workers and the connections.
// Consider that 100 workers are supposed to be running and blast needs to create 25 connections.
// This configuration will end up sharing a single connection with four workers.
// runWorkers also starts the report.ResponseReader to read from the connection,
// if it is configured to do so.
func (group *WorkerGroup) runWorkers(loadGenerationResponseChannel chan report.LoadGenerationResponse) {
	//creates new instance of Workers.
	instantiateWorkers := func() []Worker {
		connectionsSharedByWorker := group.options.concurrency / group.options.connections

		var connection net.Conn
		var err error

		var connectionId = -1
		var workers []Worker
		for count := 0; count < int(group.options.concurrency); count++ {
			if count%int(connectionsSharedByWorker) == 0 || connection == nil {
				connection, err = group.newConnection()
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "[WorkerGroup] %v\n", err.Error())
				} else {
					connectionId = connectionId + 1
				}
				if group.responseReader != nil && connection != nil {
					group.responseReader.StartReading(connection)
				}
			}
			workers = append(workers, group.instantiateWorker(connection, connectionId, loadGenerationResponseChannel))
		}
		return workers
	}

	//runs all the workers.
	//runWorkersAndWait will wait till all the workers are done.
	runWorkersAndWait := func(workers []Worker) {
		var wg sync.WaitGroup
		wg.Add(len(workers))

		for _, worker := range workers {
			worker.run(&wg)
		}
		wg.Wait()
	}
	runWorkersAndWait(instantiateWorkers())
	group.doneChannel <- struct{}{}
}

// WaitTillDone waits till all the workers are done.
func (group *WorkerGroup) WaitTillDone() {
	<-group.doneChannel
}

// DoneChannel returns the doneChannel.
func (group *WorkerGroup) DoneChannel() chan struct{} {
	return group.doneChannel
}

// newConnection creates a new TCP connection.
func (group *WorkerGroup) newConnection() (net.Conn, error) {
	connection, err := net.DialTimeout(
		"tcp",
		group.options.targetAddress,
		group.options.dialTimeout,
	)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

// instantiateWorker creates a new Worker.
func (group *WorkerGroup) instantiateWorker(connection net.Conn, connectionId int, loadGenerationResponseChannel chan report.LoadGenerationResponse) Worker {
	return Worker{
		connection:   connection,
		connectionId: connectionId,
		requestId:    group.requestId,
		options: WorkerOptions{
			maxDuration:            group.options.maxDuration,
			payloadGenerator:       group.options.payloadGenerator,
			targetAddress:          group.options.targetAddress,
			requestsPerSecond:      group.options.requestsPerSecond,
			stopChannel:            group.stopChannel,
			loadGenerationResponse: loadGenerationResponseChannel,
		},
	}
}
