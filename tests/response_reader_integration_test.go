package tests

import (
	"net"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/SarthakMakhija/blast-core/report"
)

func TestReadsResponseFromASingleConnection(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:9090", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)

	connection := connectTo(t, "localhost:9090")
	writeTo(t, connection, []byte("HelloWorld"))

	responseChannel := make(chan report.SubjectServerResponse)

	defer func() {
		server.stop()
		close(responseChannel)
		_ = connection.Close()
	}()

	responseReader := report.NewResponseReader(
		payloadSizeBytes,
		100*time.Millisecond,
		responseChannel,
	)
	responseReader.StartReading(connection)

	response := <-responseChannel

	assert.Nil(t, response.Err)
	assert.Equal(t, int64(10), response.PayloadLengthBytes)
}

func TestReadsResponseFromTwoConnections(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:9091", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)

	connection, otherConnection := connectTo(t, "localhost:9091"), connectTo(t, "localhost:9091")
	writeTo(t, connection, []byte("HelloWorld"))
	writeTo(t, otherConnection, []byte("BlastWorld"))

	time.Sleep(10 * time.Millisecond)
	responseChannel := make(chan report.SubjectServerResponse)

	defer func() {
		server.stop()
		close(responseChannel)
		_ = connection.Close()
		_ = otherConnection.Close()
	}()

	responseReader := report.NewResponseReader(
		payloadSizeBytes,
		100*time.Millisecond,
		responseChannel,
	)
	responseReader.StartReading(connection)
	responseReader.StartReading(otherConnection)

	responsesLength := captureTwoResponses(t, responseChannel)
	assert.Equal(t, int64(10), responsesLength[0])
	assert.Equal(t, int64(10), responsesLength[1])
}

func TestTracksTheNumberOfResponsesRead(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:9092", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)

	connection := connectTo(t, "localhost:9092")
	writeTo(t, connection, []byte("HelloWorld"))

	responseChannel := make(chan report.SubjectServerResponse)

	defer func() {
		server.stop()
		close(responseChannel)
		_ = connection.Close()
	}()

	responseReader := report.NewResponseReader(
		payloadSizeBytes,
		100*time.Millisecond,
		responseChannel,
	)
	responseReader.StartReading(connection)

	_ = <-responseChannel
	assert.Equal(t, uint64(1), responseReader.TotalResponsesRead())
	assert.Equal(t, uint64(1), responseReader.TotalSuccessfulResponsesRead())
}

func connectTo(t *testing.T, address string) net.Conn {
	connection, err := net.Dial("tcp", address)
	assert.Nil(t, err)

	return connection
}

func writeTo(t *testing.T, connection net.Conn, payload []byte) {
	_, err := connection.Write(payload)
	assert.Nil(t, err)
}

func captureTwoResponses(t *testing.T, responseChannel chan report.SubjectServerResponse) []int64 {
	var responsesLength []int64

	response := <-responseChannel
	assert.Nil(t, response.Err)
	responsesLength = append(responsesLength, response.PayloadLengthBytes)

	response = <-responseChannel
	assert.Nil(t, response.Err)
	responsesLength = append(responsesLength, response.PayloadLengthBytes)

	sort.Slice(responsesLength, func(i, j int) bool {
		return responsesLength[i] <= responsesLength[j]
	})

	return responsesLength
}
