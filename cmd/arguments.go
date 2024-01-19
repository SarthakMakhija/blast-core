package blast

import (
	"blast-core/payload"
)

// CommandLineArguments defines the command line arguments supported by blast.
type CommandLineArguments struct{}

// NewCommandArguments creates a new instance of CommandLineArguments.
func NewCommandArguments() CommandLineArguments {
	return CommandLineArguments{}
}

// Parse parses command line arguments using ConstantPayloadArgumentsParser.
func (arguments CommandLineArguments) Parse() Blast {
	return NewConstantPayloadArgumentsParser().Parse()
}

// ParseWithDynamicPayload parses command line arguments using DynamicPayloadArgumentsParser.
func (arguments CommandLineArguments) ParseWithDynamicPayload(payloadGenerator payload.PayloadGenerator) Blast {
	return NewDynamicPayloadArgumentsParser(payloadGenerator).Parse()
}
