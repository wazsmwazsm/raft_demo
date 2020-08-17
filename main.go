package main

import (
	"github.com/wazsmwazsm/raft_demo/rfcache"
	"log"
)

func main() {
	srv := rfcache.NewServer()

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
