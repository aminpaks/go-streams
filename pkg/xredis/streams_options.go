package xredis

// StreamConsumerOptions contains details of how steram consumer should be running
type StreamConsumerOptions struct {
	Counts  uint // amount of the consumers
	Retries int  // amount of retries for consumer entries processing
}

func (sco *StreamConsumerOptions) Normalize() {
	if sco.Counts < 1 {
		sco.Counts = 1
	}
}

func NewStreamConsumerOptions(counts uint, retries int) *StreamConsumerOptions {
	return &StreamConsumerOptions{
		Counts:  counts,
		Retries: retries,
	}
}
