package payload

// PayloadGenerator defines a function for generating the request payload
type PayloadGenerator interface {
	Generate(requestId uint64) []byte
}

// ConstantPayloadGenerator provides a constant payload to all the workers for sending the payload.
type ConstantPayloadGenerator struct {
	payload []byte
}

// NewConstantPayloadGenerator creates a new instance of ConstantPayloadGenerator.
func NewConstantPayloadGenerator(payload []byte) ConstantPayloadGenerator {
	return ConstantPayloadGenerator{
		payload: payload,
	}
}

// Generate generates (/returns) the same payload for each request.
func (generator ConstantPayloadGenerator) Generate(id uint64) []byte {
	return generator.payload
}
