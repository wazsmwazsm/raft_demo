package rfcache

import (
	"flag"
)

// Options for server
type Options struct {
	DataDir   string
	Addr      string
	APIPort   int
	RaftPort  int
	Bootstrap bool
}

// NewOptions from cli
func NewOptions() *Options {

	var node = flag.String("node", "node1", "raft node name")
	var addr = flag.String("addr", "127.0.0.1", "server addr")
	var apiPort = flag.Int("api_port", 7000, "api port")
	var raftPort = flag.Int("raft_port", 7100, "raft port")
	var bootstrap = flag.Bool("bootstrap", false, "start as raft cluster")

	return &Options{
		DataDir:   "./" + *node,
		Addr:      *addr,
		APIPort:   *apiPort,
		RaftPort:  *raftPort,
		Bootstrap: *bootstrap,
	}
}
