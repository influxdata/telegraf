package consumer

// Checkpoint interface used track consumer progress in the stream
type Checkpoint interface {
	Get(streamName, shardID string) (string, error)
	Set(streamName, shardID, sequenceNumber string) error
}

// noopCheckpoint implements the checkpoint interface with discard
type noopCheckpoint struct{}

func (n noopCheckpoint) Set(string, string, string) error   { return nil }
func (n noopCheckpoint) Get(string, string) (string, error) { return "", nil }
