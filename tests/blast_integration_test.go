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

	concurrency := uint(10)
	groupOptions := workers.NewGroupOptionsFullyLoaded(
		concurrency,
		1,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10001",
		3*time.Second,
		10,
		2*time.Second,
	)
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, false)
	blastInstance.WaitForCompletion()

	assert.True(t, extract("TotalConnections:", regexp.MustCompile("TotalConnections.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("TotalRequests:", regexp.MustCompile("TotalRequests.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("SuccessCount:", regexp.MustCompile("SuccessCount.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("ErrorCount:", regexp.MustCompile("ErrorCount.*"), buffer.Bytes()) == 0)
}

func TestBlastWithLoadGenerationForMaximumDuration(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10002", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency := uint(1000)
	groupOptions := workers.NewGroupOptionsFullyLoaded(
		concurrency,
		10,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10002",
		3*time.Second,
		10,
		2*time.Second,
	)
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, false)
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

	assert.True(t, totalRequestsMade > 1)
}

func TestBlastWithLoadGenerationAndResponseReading(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10003", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(10), uint(20)
	groupOptions := workers.NewGroupOptionsFullyLoaded(
		concurrency,
		1,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10003",
		3*time.Second,
		10,
		2*time.Second,
	)
	responseOptions := blast.ResponseOptions{
		ResponsePayloadSizeBytes: payloadSizeBytes,
		TotalResponsesToRead:     totalRequests,
		ReadingOption:            blast.ReadTotalResponses,
		ReadDeadline:             100 * time.Millisecond,
	}
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithResponseReading(groupOptions, responseOptions, false)
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "ResponseMetrics"))
	assert.True(t, extract("TotalResponses:", regexp.MustCompile("TotalResponses.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("TotalConnections:", regexp.MustCompile("TotalConnections.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("SuccessCount:", regexp.MustCompile("SuccessCount.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("ErrorCount:", regexp.MustCompile("ErrorCount.*"), buffer.Bytes()) == 0)
	assert.True(t, extract("TotalConnections:", regexp.MustCompile("TotalConnections.*"), buffer.Bytes()) >= 1)
}

func TestBlastWithLoadGenerationAndResponseReadingForMaximumDuration(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10004", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(1000), uint(1000000)
	groupOptions := workers.NewGroupOptionsFullyLoaded(
		concurrency,
		10,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10004",
		3*time.Second,
		10,
		2*time.Second,
	)
	responseOptions := blast.ResponseOptions{
		ResponsePayloadSizeBytes: payloadSizeBytes,
		TotalResponsesToRead:     totalRequests,
		ReadingOption:            blast.ReadTotalResponses,
		ReadDeadline:             100 * time.Millisecond,
	}
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithResponseReading(groupOptions, responseOptions, false)
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

	assert.True(t, totalRequestsMade > 1)
}

func TestBlastWithResponseReadingGivenTheTargetServerFailsInSendingResponses(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServerWithNoWriteback("tcp", "localhost:10005", payloadSizeBytes, 2)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency := uint(10)
	groupOptions := workers.NewGroupOptionsFullyLoaded(
		concurrency,
		1,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10005",
		3*time.Second,
		10,
		5*time.Second,
	)
	responseOptions := blast.ResponseOptions{
		ResponsePayloadSizeBytes: payloadSizeBytes,
		TotalResponsesToRead:     100,
		ReadingOption:            blast.ReadTotalResponses,
		ReadDeadline:             100 * time.Millisecond,
	}
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithResponseReading(groupOptions, responseOptions, false)
	blastInstance.WaitForCompletion()

	output := string(buffer.Bytes())

	assert.True(t, strings.Contains(output, "ResponseMetrics"))
	assert.True(t, strings.Contains(output, "i/o timeout"))
	assert.True(t, extract("TotalConnections:", regexp.MustCompile("TotalConnections.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("TotalResponses:", regexp.MustCompile("TotalResponses.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("SuccessCount:", regexp.MustCompile("SuccessCount.*"), buffer.Bytes()) >= 1)
}

func TestBlastWithLoadGenerationAndAStopSignal(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10006", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency := uint(1000)

	groupOptions := workers.NewGroupOptionsFullyLoaded(
		concurrency,
		1,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10006",
		3*time.Second,
		10,
		2*time.Second,
	)
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, false)
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

	groupOptions := workers.NewGroupOptionsFullyLoaded(
		concurrency,
		10,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10007",
		3*time.Second,
		10,
		2*time.Second,
	)
	responseOptions := blast.ResponseOptions{
		ResponsePayloadSizeBytes: payloadSizeBytes,
		TotalResponsesToRead:     totalRequests,
		ReadingOption:            blast.ReadTotalResponses,
		ReadDeadline:             100 * time.Millisecond,
	}
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithResponseReading(groupOptions, responseOptions, false)
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

	concurrency := uint(10)

	groupOptions := workers.NewGroupOptionsFullyLoaded(
		concurrency,
		1,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10008",
		3*time.Second,
		10,
		2*time.Second,
	)
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, true)

	go func() {
		time.Sleep(3 * time.Second)
		blastInstance.Stop()
	}()

	blastInstance.WaitForCompletion()

	assert.True(t, extract("TotalConnections:", regexp.MustCompile("TotalConnections.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("TotalRequests:", regexp.MustCompile("TotalRequests.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("SuccessCount:", regexp.MustCompile("SuccessCount.*"), buffer.Bytes()) >= 1)
	assert.True(t, extract("ErrorCount:", regexp.MustCompile("ErrorCount.*"), buffer.Bytes()) == 0)
}

func TestBlastWithLoadGenerationForMaximumDurationAndConnectionsAreKeptAlive(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10009", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency := uint(1000)
	groupOptions := workers.NewGroupOptionsWithConnections(
		concurrency,
		10,
		payload.NewConstantPayloadGenerator([]byte("HelloWorld")),
		"localhost:10009",
	)
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer

	blastInstance := blast.NewBlastWithoutResponseReading(groupOptions, true)
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

func extract(textToFind string, regularExpression *regexp.Regexp, buffer []byte) int {
	found := regularExpression.Find(buffer)
	asInt, _ := strconv.Atoi(strings.Trim(
		strings.ReplaceAll(string(found), textToFind, ""),
		" ",
	))
	return asInt
}
