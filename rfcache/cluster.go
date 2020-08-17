package rfcache

import (
	"fmt"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"net"
	"os"
	"path/filepath"
	"time"
)

// RaftNode for cluster
type RaftNode struct {
	raft           *raft.Raft
	fsm            *FSM
	leaderNotifyCh chan bool
}

func newRaftTransport(opts *Options) (*raft.NetworkTransport, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", opts.Addr, opts.RaftPort))
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(addr.String(), addr, 5, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	return transport, nil
}

// NewRaftNode create raft node
func NewRaftNode(opts *Options, cache *Cache) (*RaftNode, error) {

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(fmt.Sprintf("%s:%d", opts.Addr, opts.RaftPort))
	raftConf.SnapshotInterval = 20 * time.Second
	raftConf.SnapshotThreshold = 2
	leaderNotifyCh := make(chan bool, 1)
	raftConf.NotifyCh = leaderNotifyCh

	trans, err := newRaftTransport(opts)
	if err != nil {
		return nil, err
	}

	fsm := NewFSM(cache)

	snapshotStore, err := raft.NewFileSnapshotStore(opts.DataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(opts.DataDir, "raft-log.bolt"))
	if err != nil {
		return nil, err
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(opts.DataDir, "raft-stable.bolt"))
	if err != nil {
		return nil, err
	}

	raftNode, err := raft.NewRaft(raftConf, fsm, logStore, stableStore, snapshotStore, trans)
	if err != nil {
		return nil, err
	}

	if opts.Bootstrap { // master
		raftNode.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				raft.Server{
					ID:      raftConf.LocalID,
					Address: trans.LocalAddr(),
				},
			},
		})
	}

	return &RaftNode{
		raft:           raftNode,
		fsm:            fsm,
		leaderNotifyCh: leaderNotifyCh,
	}, nil
}
