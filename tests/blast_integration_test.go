package tests

import (
	"bytes"
	blast "github.com/SarthakMakhija/blast-core/cmd"
	"github.com/SarthakMakhija/blast-core/payload"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/SarthakMakhija/blast-core/workers"
)

func TestBlastWithLoadGeneration(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10001", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(10), uint(20)

	groupOptions := workers.NewGroupOptions(concurrency, totalRequests, 1, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:10001")
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, 5*time.Minute, false)
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalConnections: 1"))
	assert.True(t, strings.Contains(output, "TotalRequests: 20"))
	assert.True(t, strings.Contains(output, "SuccessCount: 20"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))
	assert.True(t, strings.Contains(output, "TotalPayloadSize: 200 B"))
	assert.True(t, strings.Contains(output, "AveragePayloadSize: 10 B"))
}

func TestBlastWithLoadGenerationForMaximumDuration(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10002", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(1000), uint(2_00_000)

	groupOptions := workers.NewGroupOptionsWithConnections(
		concurrency,
		10,
		totalRequests,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10002",
	)
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, 10*time.Millisecond, false)
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalRequests"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))

	regularExpression := regexp.MustCompile("TotalRequests.*")
	totalRequestsString := regularExpression.Find(buffer.Bytes())
	totalRequestsMade, _ := strconv.Atoi(strings.Trim(
		strings.ReplaceAll(string(totalRequestsString), "TotalRequests:", ""),
		" ",
	))

	assert.True(t, totalRequestsMade < 2_00_000)
}

func TestBlastWithLoadGenerationAndResponseReading(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10003", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(10), uint(20)

	groupOptions := workers.NewGroupOptions(concurrency, totalRequests, 1, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:10003")
	responseOptions := blast.ResponseOptions{
		ResponsePayloadSizeBytes: payloadSizeBytes,
		TotalResponsesToRead:     totalRequests,
		ReadingOption:            blast.ReadTotalResponses,
		ReadDeadline:             100 * time.Millisecond,
	}
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithResponseReading(groupOptions, responseOptions, 5*time.Minute, false)
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "ResponseMetrics"))
	assert.True(t, strings.Contains(output, "TotalResponses: 20"))
	assert.True(t, strings.Contains(output, "TotalConnections: 1"))
	assert.True(t, strings.Contains(output, "SuccessCount: 20"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))
	assert.True(t, strings.Contains(output, "TotalPayloadSize: 200 B"))
	assert.True(t, strings.Contains(output, "AveragePayloadSize: 10 B"))
}

func TestBlastWithLoadGenerationAndResponseReadingForMaximumDuration(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10004", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(1000), uint(2_00_000)

	groupOptions := workers.NewGroupOptionsWithConnections(
		concurrency,
		10,
		totalRequests,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10004",
	)
	responseOptions := blast.ResponseOptions{
		ResponsePayloadSizeBytes: payloadSizeBytes,
		TotalResponsesToRead:     totalRequests,
		ReadingOption:            blast.ReadTotalResponses,
		ReadDeadline:             100 * time.Millisecond,
	}
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithResponseReading(groupOptions, responseOptions, 10*time.Millisecond, false)
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalRequests"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))
	assert.True(t, strings.Contains(output, "ResponseMetrics"))

	regularExpression := regexp.MustCompile("TotalRequests.*")
	totalRequestsString := regularExpression.Find(buffer.Bytes())
	totalRequestsMade, _ := strconv.Atoi(strings.Trim(
		strings.ReplaceAll(string(totalRequestsString), "TotalRequests:", ""),
		" ",
	))

	assert.True(t, totalRequestsMade < 2_00_000)
}

func TestBlastWithResponseReadingGivenTheTargetServerFailsInSendingResponses(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServerWithNoWriteback("tcp", "localhost:10005", payloadSizeBytes, 2)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(10), uint(20)

	groupOptions := workers.NewGroupOptions(concurrency, totalRequests, 1, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:10005")
	responseOptions := blast.ResponseOptions{
		ResponsePayloadSizeBytes: payloadSizeBytes,
		TotalResponsesToRead:     20,
		ReadingOption:            blast.ReadTotalResponses,
		ReadDeadline:             100 * time.Millisecond,
	}
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithResponseReading(groupOptions, responseOptions, 5*time.Second, false)
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())

	assert.True(t, strings.Contains(output, "ResponseMetrics"))
	assert.True(t, strings.Contains(output, "TotalResponses: 20"))
	assert.True(t, strings.Contains(output, "SuccessCount: 10"))
	assert.True(t, strings.Contains(output, "ErrorCount: 10"))
	assert.True(t, strings.Contains(output, "TotalResponsePayloadSize: 100 B"))
	assert.True(t, strings.Contains(output, "AveragePayloadSize: 10 B"))
}

func TestBlastWithLoadGenerationAndAStopSignal(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10006", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(1000), uint(2_00_000)

	groupOptions := workers.NewGroupOptions(concurrency, totalRequests, 1, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:10006")
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, 50*time.Second, false)
	go func() {
		time.Sleep(10 * time.Millisecond)
		blastInstance.Stop()
	}()
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalRequests"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))

	regularExpression := regexp.MustCompile("TotalRequests.*")
	totalRequestsString := regularExpression.Find(buffer.Bytes())
	totalRequestsMade, _ := strconv.Atoi(strings.Trim(
		strings.ReplaceAll(string(totalRequestsString), "TotalRequests:", ""),
		" ",
	))

	assert.True(t, totalRequestsMade < 2_00_000)
}

func TestBlastWithLoadGenerationAndResponseReadingWithStopSignal(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10007", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(1000), uint(2_00_000)

	groupOptions := workers.NewGroupOptionsWithConnections(
		concurrency,
		10,
		totalRequests,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10007",
	)
	responseOptions := blast.ResponseOptions{
		ResponsePayloadSizeBytes: payloadSizeBytes,
		TotalResponsesToRead:     totalRequests,
		ReadingOption:            blast.ReadTotalResponses,
		ReadDeadline:             100 * time.Millisecond,
	}
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithResponseReading(groupOptions, responseOptions, 50*time.Millisecond, false)
	go func() {
		time.Sleep(10 * time.Millisecond)
		blastInstance.Stop()
	}()
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalRequests"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))
	assert.True(t, strings.Contains(output, "ResponseMetrics"))

	regularExpression := regexp.MustCompile("TotalRequests.*")
	totalRequestsString := regularExpression.Find(buffer.Bytes())
	totalRequestsMade, _ := strconv.Atoi(strings.Trim(
		strings.ReplaceAll(string(totalRequestsString), "TotalRequests:", ""),
		" ",
	))

	assert.True(t, totalRequestsMade < 2_00_000)
}

func TestBlastWithLoadGenerationAndConnectionsAreKeptAlive(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10008", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(10), uint(20)

	groupOptions := workers.NewGroupOptions(concurrency, totalRequests, 1, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:10008")
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, 5*time.Minute, true)

	go func() {
		time.Sleep(200 * time.Millisecond)
		blastInstance.Stop()
	}()

	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalConnections: 1"))
	assert.True(t, strings.Contains(output, "TotalRequests: 20"))
	assert.True(t, strings.Contains(output, "SuccessCount: 20"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))
	assert.True(t, strings.Contains(output, "TotalPayloadSize: 200 B"))
	assert.True(t, strings.Contains(output, "AveragePayloadSize: 10 B"))
}

func TestBlastWithLoadGenerationForMaximumDurationAndConnectionsAreKeptAlive(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10009", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(1000), uint(2_00_000)
	groupOptions := workers.NewGroupOptionsWithConnections(
		concurrency,
		10,
		totalRequests,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10009",
	)
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, 200*time.Millisecond, true)
	go func() {
		time.Sleep(120 * time.Millisecond)
		blastInstance.Stop()
	}()

	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalRequests"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))

	regularExpression := regexp.MustCompile("TotalRequests.*")
	totalRequestsString := regularExpression.Find(buffer.Bytes())
	totalRequestsMade, _ := strconv.Atoi(strings.Trim(
		strings.ReplaceAll(string(totalRequestsString), "TotalRequests:", ""),
		" ",
	))

	assert.True(t, totalRequestsMade < 2_00_000)
}

func TestBlastWithLoadGenerationRepeatingNTimes(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10010", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests, repeat := uint(10), uint(20), uint(4)

	groupOptions := workers.NewGroupOptions(concurrency, totalRequests, repeat, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:10010")
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, 5*time.Minute, false)
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalConnections: 1"))
	assert.True(t, strings.Contains(output, "TotalRequests: 80"))
	assert.True(t, strings.Contains(output, "SuccessCount: 80"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))
	assert.True(t, strings.Contains(output, "TotalPayloadSize: 800 B"))
	assert.True(t, strings.Contains(output, "AveragePayloadSize: 10 B"))
}

func TestBlastWithLoadGenerationRepeatingNTimesAndResponseReading(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10011", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests, repeat := uint(10), uint(20), uint(5)

	groupOptions := workers.NewGroupOptions(concurrency, totalRequests, repeat, payload.NewConstantPayloadGenerator([]byte("HelloWorld")), "localhost:10011")
	responseOptions := blast.ResponseOptions{
		ResponsePayloadSizeBytes: payloadSizeBytes,
		TotalResponsesToRead:     totalRequests * repeat,
		ReadingOption:            blast.ReadTotalResponses,
		ReadDeadline:             100 * time.Millisecond,
	}
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithResponseReading(groupOptions, responseOptions, 5*time.Minute, false)
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "ResponseMetrics"))
	assert.True(t, strings.Contains(output, "TotalResponses: 100"))
	assert.True(t, strings.Contains(output, "TotalConnections: 1"))
	assert.True(t, strings.Contains(output, "SuccessCount: 100"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))
	assert.True(t, strings.Contains(output, "TotalPayloadSize: 1.0 kB"))
	assert.True(t, strings.Contains(output, "AveragePayloadSize: 10 B"))
}
