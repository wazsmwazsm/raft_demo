package main

import (
	"github.com/wazsmwazsm/raft_demo/rfcache"
	"log"
)

func main() {
	srv, err := rfcache.NewServer(rfcache.NewOptions())
	if err != nil {
		log.Fatal(err)
	}

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
