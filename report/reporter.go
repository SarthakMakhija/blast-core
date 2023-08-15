package report

import (
	"time"

	"blast/workers"
)

type Report struct {
	Load     LoadMetrics
	Response ResponseMetrics
}

// TODO: Total connections, total requests, requests per second
type LoadMetrics struct {
	SuccessCount              uint
	ErrorCount                uint
	ErrorCountByType          map[string]uint
	TotalPayloadLengthBytes   int64
	AveragePayloadLengthBytes float64
	EarliestLoadSendTime      time.Time
	LatestLoadSendTime        time.Time
}

// TODO: connection time, total responses
type ResponseMetrics struct {
	SuccessCount                      uint
	ErrorCount                        uint
	ErrorCountByType                  map[string]uint
	TotalResponsePayloadLengthBytes   int64
	AverageResponsePayloadLengthBytes float64
	EarliestResponseReceivedTime      time.Time
	LatestResponseReceivedTime        time.Time
}

type Reporter struct {
	report                *Report
	loadGenerationChannel chan workers.LoadGenerationResponse
	responseChannel       chan SubjectServerResponse
}

func NewReporter(
	loadGenerationChannel chan workers.LoadGenerationResponse,
	responseChannel chan SubjectServerResponse,
) Reporter {
	return Reporter{
		report: &Report{
			Load: LoadMetrics{
				ErrorCountByType: make(map[string]uint),
			},
			Response: ResponseMetrics{
				ErrorCountByType: make(map[string]uint),
			},
		},
		loadGenerationChannel: loadGenerationChannel,
		responseChannel:       responseChannel,
	}
}

func (reporter *Reporter) Run() {
	reporter.collectLoadMetrics()
	reporter.collectResponseMetrics()
}

func (reporter *Reporter) collectLoadMetrics() {
	go func() {
		totalGeneratedLoad := 0
		for load := range reporter.loadGenerationChannel {
			totalGeneratedLoad++

			if load.Err != nil {
				reporter.report.Load.ErrorCount++
				reporter.report.Load.ErrorCountByType[load.Err.Error()]++
			} else {
				reporter.report.Load.SuccessCount++
			}
			reporter.report.Load.TotalPayloadLengthBytes += load.PayloadLength
			reporter.report.Load.AveragePayloadLengthBytes = float64(
				reporter.report.Load.TotalPayloadLengthBytes,
			) / float64(
				totalGeneratedLoad,
			)

			if reporter.report.Load.EarliestLoadSendTime.IsZero() ||
				load.LoadGenerationTime.Before(reporter.report.Load.EarliestLoadSendTime) {
				reporter.report.Load.EarliestLoadSendTime = load.LoadGenerationTime
			}

			if reporter.report.Load.LatestLoadSendTime.IsZero() ||
				load.LoadGenerationTime.After(reporter.report.Load.LatestLoadSendTime) {
				reporter.report.Load.LatestLoadSendTime = load.LoadGenerationTime
			}
		}
	}()
}

func (reporter *Reporter) collectResponseMetrics() {
	go func() {
		totalResponses := 0
		for response := range reporter.responseChannel {
			totalResponses++

			if response.Err != nil {
				reporter.report.Response.ErrorCount++
				reporter.report.Response.ErrorCountByType[response.Err.Error()]++
			} else {
				reporter.report.Response.SuccessCount++
			}
			reporter.report.Response.TotalResponsePayloadLengthBytes += response.PayloadLength
			reporter.report.Response.AverageResponsePayloadLengthBytes = float64(
				reporter.report.Response.TotalResponsePayloadLengthBytes,
			) / float64(totalResponses)

			if reporter.report.Response.EarliestResponseReceivedTime.IsZero() ||
				response.ResponseTime.Before(
					reporter.report.Response.EarliestResponseReceivedTime,
				) {
				reporter.report.Response.EarliestResponseReceivedTime = response.ResponseTime
			}

			if reporter.report.Response.LatestResponseReceivedTime.IsZero() ||
				response.ResponseTime.After(
					reporter.report.Response.LatestResponseReceivedTime,
				) {
				reporter.report.Response.LatestResponseReceivedTime = response.ResponseTime
			}
		}
	}()
}
