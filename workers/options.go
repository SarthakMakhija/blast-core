package workers

type GroupOptions struct {
	concurrency       uint
	connections       uint
	totalRequests     uint
	payload           []byte
	targetAddress     string
	requestsPerSecond float64
}

type WorkerOptions struct {
	totalRequests     uint
	payload           []byte
	targetAddress     string
	requestsPerSecond float64
	stopChannel       chan struct{}
	responseChannel   chan WorkerResponse
}

// TODO: Rename this
type WorkerResponse struct {
	Err           error
	PayloadLength int64
}

func NewGroupOptions(
	concurrency uint,
	totalRequests uint,
	payload []byte,
	targetAddress string,
) GroupOptions {
	return NewGroupOptionsFullyLoaded(
		concurrency,
		1,
		totalRequests,
		payload,
		targetAddress,
		0.0,
	)
}

func NewGroupOptionsWithConnections(
	concurrency uint,
	connections uint,
	totalRequests uint,
	payload []byte,
	targetAddress string,
) GroupOptions {
	return NewGroupOptionsFullyLoaded(
		concurrency,
		connections,
		totalRequests,
		payload,
		targetAddress,
		0.0,
	)
}

func NewGroupOptionsFullyLoaded(
	concurrency uint,
	connections uint,
	totalRequests uint,
	payload []byte,
	targetAddress string,
	requestsPerSecond float64,
) GroupOptions {
	return GroupOptions{
		concurrency:       concurrency,
		connections:       connections,
		totalRequests:     totalRequests,
		payload:           payload,
		targetAddress:     targetAddress,
		requestsPerSecond: requestsPerSecond,
	}
}