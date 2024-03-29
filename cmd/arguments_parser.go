package blast

import (
	"flag"
	"fmt"
	"github.com/SarthakMakhija/blast-core/payload"
	"github.com/SarthakMakhija/blast-core/workers"
	"github.com/dimiro1/banner"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	concurrency             = flag.Uint("c", 50, "")
	connections             = flag.Uint("conn", 1, "")
	keepConnectionsAlive    = flag.Bool("kA", false, "")
	payloadFilePath         = flag.String("f", "", "")
	requestsPerSecond       = flag.Float64("rps", 50, "")
	maxDuration             = flag.Duration("z", 20*time.Second, "")
	connectTimeout          = flag.Duration("t", 3*time.Second, "")
	readResponses           = flag.Bool("Rr", false, "")
	responsePayloadSize     = flag.Int64("Rrs", -1, "")
	readResponseDeadline    = flag.Duration("Rrd", 0*time.Second, "")
	readTotalResponses      = flag.Uint("Rtr", 0, "")
	readSuccessfulResponses = flag.Uint("Rsr", 0, "")
	cpus                    = flag.Int("cpus", runtime.GOMAXPROCS(-1), "")
)

var exitFunction = usageAndExit

// ArgumentsParser supports parsing command line arguments.
type ArgumentsParser interface {
	Parse(executableName string) Blast
}

// ConstantPayloadArgumentsParser parses command line arguments with constant payload.
type ConstantPayloadArgumentsParser struct{}

// DynamicPayloadArgumentsParser parses command line arguments with dynamic payload.
type DynamicPayloadArgumentsParser struct {
	payloadGenerator payload.PayloadGenerator
}

// NewConstantPayloadArgumentsParser creates a new instance of ConstantPayloadArgumentsParser.
func NewConstantPayloadArgumentsParser() ConstantPayloadArgumentsParser {
	return ConstantPayloadArgumentsParser{}
}

// NewDynamicPayloadArgumentsParser creates a new instance of DynamicPayloadArgumentsParser.
func NewDynamicPayloadArgumentsParser(payloadGenerator payload.PayloadGenerator) DynamicPayloadArgumentsParser {
	return DynamicPayloadArgumentsParser{
		payloadGenerator: payloadGenerator,
	}
}

// Parse parses the command line arguments.
func (parser ConstantPayloadArgumentsParser) Parse(executableName string) Blast {
	logo := `{{ .Title "%v" "" 0}}`
	banner.InitString(os.Stdout, true, false, fmt.Sprintf(logo, executableName))

	flag.Usage = func() {
		var usage = `%v is a load generator for TCP servers which maintain persistent connections.

Usage: %v [options...] <url>

Options:
  -c      Number of workers to run concurrently. Default is 50.
  -f      File path containing the load payload.
  -rps    Rate limit in requests per second (RPS) per worker. Default is 50.
  -z      Duration of blast to send requests. When duration is reached,
          application stops and exits. Default is 20 seconds.
          Example usage: -z 10s or -z 3m.
  -t      Timeout for establishing connection with the target server. Default is 3 seconds.
          Also called as DialTimeout.
  -Rr     Read responses from the target server. Default is false.
  -Rrs    Read response size is the size of the responses in bytes returned by the target server. 
  -Rrd    Read response deadline defines the deadline for the read calls on connection.
          Default is no deadline which means the read calls do not timeout.
          This flag is applied only if "Read responses" (-Rr) is true.
  -Rtr    Read total responses is the total responses to read from the target server. 
          blast will stop if either the duration (-z) has exceeded or the total 
          responses have been read. This flag is applied only if "Read responses" (-Rr)
          is true.
  -Rsr    Read successful responses is the total successful responses to read from the target server. 
          blast will stop if either the duration (-z) has exceeded or 
          the total successful responses have been read. Either of "-Rtr"
          or "-Rsr" must be specified, if -Rr is set. This flag is applied only if 
          "Read responses" (-Rr) is true.

  -conn   Number of connections to open with the target URL.
          Total number of connections cannot be greater than the concurrency level.
          Also, concurrency level modulo connections must be equal to zero.
          Default is 1.

  -kA     Keep connections alive. If set, blast will keep running until a termination signal is sent. Default is false.

  -cpus   Number of cpu cores to use.
          (default for current machine is %d cores)
`
		_, _ = fmt.Fprint(os.Stderr, fmt.Sprintf(usage, executableName, executableName, runtime.NumCPU()))
	}

	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit("")
	}

	url := flag.Args()[0]
	assertUrl(url)
	assertPayloadFilePath(*payloadFilePath)
	assertConnectTimeout(*connectTimeout)
	assertRequestsPerSecond(*requestsPerSecond)
	assertMaxDuration(*maxDuration)
	assertConcurrencyWithClientConnections(
		*concurrency,
		*connections,
	)
	assertResponseReading(
		*readResponses,
		*responsePayloadSize,
		*readTotalResponses,
		*readSuccessfulResponses,
	)
	assertAndSetMaxProcs(*cpus)
	return setUpBlast(
		payload.NewConstantPayloadGenerator(getFilePayload(*payloadFilePath)),
		url,
	)
}

// Parse parses the command line arguments.
func (parser DynamicPayloadArgumentsParser) Parse(executableName string) Blast {
	logo := `{{ .Title "%v" "" 0}}`
	banner.InitString(os.Stdout, true, false, fmt.Sprintf(logo, executableName))

	flag.Usage = func() {
		var usage = `%v is a load generator for TCP servers which maintain persistent connections.

Usage: %v [options...] <url>

Options:
  -c      Number of workers to run concurrently. Default is 50.
  -rps    Rate limit in requests per second (RPS) per worker. Default is 50.
  -z      Duration of blast to send requests. When duration is reached,
          application stops and exits. Default is 20 seconds.
          Example usage: -z 10s or -z 3m.
  -t      Timeout for establishing connection with the target server. Default is 3 seconds.
          Also called as DialTimeout.
  -Rr     Read responses from the target server. Default is false.
  -Rrs    Read response size is the size of the responses in bytes returned by the target server. 
  -Rrd    Read response deadline defines the deadline for the read calls on connection.
          Default is no deadline which means the read calls do not timeout.
          This flag is applied only if "Read responses" (-Rr) is true.
  -Rtr    Read total responses is the total responses to read from the target server. 
          blast will stop if either the duration (-z) has exceeded or the total 
          responses have been read. This flag is applied only if "Read responses" (-Rr)
          is true.
  -Rsr    Read successful responses is the total successful responses to read from the target server. 
          blast will stop if either the duration (-z) has exceeded or 
          the total successful responses have been read. Either of "-Rtr"
          or "-Rsr" must be specified, if -Rr is set. This flag is applied only if 
          "Read responses" (-Rr) is true.

  -conn   Number of connections to open with the target URL.
          Total number of connections cannot be greater than the concurrency level.
          Also, concurrency level modulo connections must be equal to zero.
          Default is 1.
	
  -kA     Keep connections alive. If set, blast will keep running until a termination signal is sent. Default is false.

  -cpus   Number of cpu cores to use.
          (default for current machine is %d cores)
`
		_, _ = fmt.Fprint(os.Stderr, fmt.Sprintf(usage, executableName, executableName, runtime.NumCPU()))
	}

	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit("")
	}

	url := flag.Args()[0]
	assertUrl(url)
	assertConnectTimeout(*connectTimeout)
	assertRequestsPerSecond(*requestsPerSecond)
	assertMaxDuration(*maxDuration)
	assertConcurrencyWithClientConnections(
		*concurrency,
		*connections,
	)
	assertResponseReading(
		*readResponses,
		*responsePayloadSize,
		*readTotalResponses,
		*readSuccessfulResponses,
	)
	assertAndSetMaxProcs(*cpus)
	return setUpBlast(parser.payloadGenerator, url)
}

// assertUrl asserts that the URL is not empty.
func assertUrl(url string) {
	if len(strings.Trim(url, " ")) == 0 {
		exitFunction("URL cannot be blank. URL is of the form host:port.")
	}
}

// assertPayloadFilePath asserts that the payloadFilePath is not empty.
func assertPayloadFilePath(filePath string) {
	if len(strings.Trim(filePath, " ")) == 0 {
		exitFunction("-f cannot be blank.")
	}
}

// assertConnectTimeout asserts that the connectTimeout is greater than zero.
func assertConnectTimeout(timeout time.Duration) {
	if timeout <= time.Duration(0) {
		exitFunction("-t cannot be smaller than or equal to zero.")
	}
}

// assertRequestsPerSecond asserts that the requestsPerSecond is greater than zero.
func assertRequestsPerSecond(requestsPerSecond float64) {
	if requestsPerSecond <= 0 {
		exitFunction("-rps cannot be smaller than or equal to zero.")
	}
}

// assertMaxDuration asserts that the maxDuration is greater than zero.
func assertMaxDuration(duration time.Duration) {
	if duration <= time.Duration(0) {
		exitFunction("-z cannot be smaller than or equal to zero.")
	}
}

// assertConcurrencyWithClientConnections asserts the relationship between concurrency and
// client connections.
func assertConcurrencyWithClientConnections(
	concurrency, connections uint,
) {
	if connections <= 0 {
		exitFunction("-conn cannot be smaller than 1.")
	}
	if concurrency <= 0 {
		exitFunction("-c cannot be smaller than 1.")
	}
	if connections > concurrency {
		exitFunction("-conn cannot be greater than -c.")
	}
	if concurrency%connections != 0 {
		exitFunction("-c modulo -conn must be equal to zero.")
	}
}

// assertAndSetMaxProcs asserts the maximum number of cpus and sets the value in GOMAXPROCS.
func assertAndSetMaxProcs(cpus int) {
	if cpus <= 0 {
		exitFunction("-cpus cannot be smaller than 1.")
	}
	runtime.GOMAXPROCS(cpus)
}

// assertResponseReading asserts the options related to reading responses.
func assertResponseReading(
	readResponses bool,
	responsePayloadSize int64,
	readTotalResponses, readSuccessfulResponses uint,
) {
	if readResponses {
		if responsePayloadSize < 0 {
			exitFunction("-Rrs cannot be smaller than 0.")
		}
		if readTotalResponses > 0 && readSuccessfulResponses > 0 {
			exitFunction("both -Rtr and -Rsr cannot be specified.")
		}
		if readTotalResponses == 0 && readSuccessfulResponses == 0 {
			exitFunction("either of -Rtr or -Rsr must be specified.")
		}
	}
}

// getFilePayload returns the file content.
func getFilePayload(filePath string) []byte {
	provider, err := payload.NewFilePayloadProvider(filePath)
	if err != nil {
		exitFunction(fmt.Sprintf("file path: %v does not exist.", filePath))
	}
	return provider.Get()
}

// usageAndExit defines the usage of blast application and exits the application.
func usageAndExit(msg string) {
	if msg != "" {
		_, _ = fmt.Fprintf(os.Stderr, msg)
		_, _ = fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

// setUpBlast creates a new instance of blast.Blast.
func setUpBlast(
	payloadGenerator payload.PayloadGenerator,
	url string,
) Blast {
	groupOptions := workers.NewGroupOptionsFullyLoaded(
		*concurrency,
		*connections,
		payloadGenerator,
		url,
		*connectTimeout,
		*requestsPerSecond,
		*maxDuration,
	)

	var instance Blast
	if *readResponses {
		readingOption := ReadTotalResponses
		if *readSuccessfulResponses > 0 {
			readingOption = ReadSuccessfulResponses
		}
		responseOptions := ResponseOptions{
			ResponsePayloadSizeBytes:       *responsePayloadSize,
			TotalResponsesToRead:           *readTotalResponses,
			TotalSuccessfulResponsesToRead: *readSuccessfulResponses,
			ReadingOption:                  readingOption,
			ReadDeadline:                   *readResponseDeadline,
		}
		instance = NewBlastWithResponseReading(groupOptions, responseOptions, *keepConnectionsAlive)
	} else {
		instance = NewBlastWithoutResponseReading(groupOptions, *keepConnectionsAlive)
	}
	return instance
}
