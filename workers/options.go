package workers

import (
	"github.com/SarthakMakhija/blast-core/payload"
	"time"

	"github.com/SarthakMakhija/blast-core/report"
)

const dialTimeout = 3 * time.Second

// GroupOptions defines the configuration options for the WorkerGroup.
type GroupOptions struct {
	concurrency       uint
	connections       uint
	payloadGenerator  payload.PayloadGenerator
	targetAddress     string
	requestsPerSecond float64
	maxDuration       time.Duration
	dialTimeout       time.Duration
}

// WorkerOptions defines the configuration options for a running Worker.
type WorkerOptions struct {
	maxDuration            time.Duration
	payloadGenerator       payload.PayloadGenerator
	targetAddress          string
	requestsPerSecond      float64
	stopChannel            chan struct{}
	loadGenerationResponse chan report.LoadGenerationResponse
}

// NewGroupOptions creates a new instance of GroupOptions.
func NewGroupOptions(
	concurrency uint,
	payloadGenerator payload.PayloadGenerator,
	targetAddress string,
	maxDuration time.Duration,
) GroupOptions {
	return NewGroupOptionsFullyLoaded(
		concurrency,
		1,
		payloadGenerator,
		targetAddress,
		dialTimeout,
		1.0,
		maxDuration,
	)
}

// NewGroupOptionsWithConnections creates a new instance of GroupOptions.
func NewGroupOptionsWithConnections(
	concurrency uint,
	connections uint,
	payloadGenerator payload.PayloadGenerator,
	targetAddress string,
) GroupOptions {
	return NewGroupOptionsFullyLoaded(
		concurrency,
		connections,
		payloadGenerator,
		targetAddress,
		dialTimeout,
		0.0,
		2*time.Millisecond,
	)
}

// NewGroupOptionsFullyLoaded creates a new instance of GroupOptions.
func NewGroupOptionsFullyLoaded(
	concurrency uint,
	connections uint,
	payloadGenerator payload.PayloadGenerator,
	targetAddress string,
	dialTimeout time.Duration,
	requestsPerSecond float64,
	maxDuration time.Duration,
) GroupOptions {
	return GroupOptions{
		concurrency:       concurrency,
		connections:       connections,
		payloadGenerator:  payloadGenerator,
		targetAddress:     targetAddress,
		requestsPerSecond: requestsPerSecond,
		maxDuration:       maxDuration,
		dialTimeout:       dialTimeout,
	}
}

// TotalRequests returns the total number of requests set in GroupOptions across all the runs defined by repeat flag.
func (groupOptions GroupOptions) TotalRequests() uint {
	return 100 //TODO: Remove this method
}

// ExpectedLoadInTotalDuration returns the expected total load.
func (groupOptions GroupOptions) ExpectedLoadInTotalDuration() uint {
	return uint(groupOptions.requestsPerSecond * float64(groupOptions.concurrency) * (groupOptions.maxDuration.Seconds()))
}

// MaxDuration returns the maximum duration.
func (groupOptions GroupOptions) MaxDuration() time.Duration {
	return groupOptions.maxDuration
}
