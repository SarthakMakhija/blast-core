package blast

import (
	"github.com/SarthakMakhija/blast-core/payload"
)

// CommandLineArguments defines the command line arguments supported by blast.
type CommandLineArguments struct{}

// NewCommandArguments creates a new instance of CommandLineArguments.
func NewCommandArguments() CommandLineArguments {
	return CommandLineArguments{}
}

// Parse parses command line arguments using ConstantPayloadArgumentsParser.
func (arguments CommandLineArguments) Parse(executableName string) Blast {
	return NewConstantPayloadArgumentsParser().Parse(executableName)
}

// ParseWithDynamicPayload parses command line arguments using DynamicPayloadArgumentsParser.
func (arguments CommandLineArguments) ParseWithDynamicPayload(executableName string, payloadGenerator payload.PayloadGenerator) Blast {
	return NewDynamicPayloadArgumentsParser(payloadGenerator).Parse(executableName)
}
