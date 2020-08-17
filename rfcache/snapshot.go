package rfcache

import (
	"github.com/hashicorp/raft"
)

type snapshot struct {
	cache *Cache
}

// Persist should dump all necessary state to the WriteCloser 'sink',
// and call sink.Close() when finished or call sink.Cancel() on error.
func (s *snapshot) Persist(sink raft.SnapshotSink) error {

	snapshotBytes, err := s.cache.Marshal()
	if err != nil {
		sink.Cancel()
		return err
	}

	if _, err := sink.Write(snapshotBytes); err != nil {
		sink.Cancel()
		return err
	}

	if err := sink.Close(); err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

// Release is invoked when we are finished with the snapshot.
func (s *snapshot) Release() {}
