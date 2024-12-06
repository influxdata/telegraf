package kinesis_consumer

import "context"

// noopStore implements the storage interface with discard
type noopStore struct{}

func (noopStore) SetCheckpoint(_ context.Context, _, _, _ string) error        { return nil }
func (noopStore) GetCheckpoint(_ context.Context, _, _ string) (string, error) { return "", nil }
