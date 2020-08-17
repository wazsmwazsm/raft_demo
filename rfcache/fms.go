package rfcache

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	"io"
)

// FSM for node
type FSM struct {
	cache *Cache
}

// NewFSM create fsm
func NewFSM(cache *Cache) *FSM {
	return &FSM{
		cache: cache,
	}
}

type logEntryData struct {
	Key   string
	Value string
}

// Apply log is invoked once a log entry is committed.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	data := &logEntryData{}

	if err := json.Unmarshal(logEntry.Data, data); err != nil {
		return err
	}

	f.cache.Set(data.Key, data.Value)

	return nil
}

// Snapshot is used to support log compaction. This call should
// return an FSMSnapshot which can be used to save a point-in-time
// snapshot of the FSM. Apply and Snapshot are not called in multiple
// threads, but Apply will be called concurrently with Persist. This means
// the FSM should be implemented in a fashion that allows for concurrent
// updates while a snapshot is happening.
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{cache: f.cache}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (f *FSM) Restore(serialized io.ReadCloser) error {

	return f.cache.Unmarshal(serialized)
}
