package rfcache

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"net/http"
	"time"
)

// Server http
type Server struct {
	httpSrv *http.Server
	node    *RaftNode
	cache   *Cache
}

// NewServer create server
func NewServer(opts *Options) (*Server, error) {

	mux := gin.New()
	cache := NewCache()
	raftNode, err := NewRaftNode(opts, cache)
	if err != nil {
		return nil, err
	}
	srv := &Server{
		httpSrv: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", opts.Addr, opts.APIPort),
			Handler: mux,
		},
		cache: cache,
		node:  raftNode,
	}
	mux.GET("/get", srv.GetCache)
	mux.POST("/set", srv.SetCache)
	mux.POST("/join", srv.Join)

	return srv, nil
}

// Run server
func (s *Server) Run() error {
	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// GetCache get cache
func (s *Server) GetCache(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "key is empty", "data": struct{}{}})
		return
	}

	value := s.cache.Get(key)

	c.JSON(200, gin.H{
		"code":           0,
		"error_messaage": "",
		"data":           value,
	})
}

type setReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SetCache set cache
func (s *Server) SetCache(c *gin.Context) {
	var req setReq
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "json parse err:" + err.Error(), "data": struct{}{}})
		return
	}

	key := req.Key
	if key == "" {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "key is empty", "data": struct{}{}})
		return
	}
	value := req.Value
	if value == "" {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "value is empty", "data": struct{}{}})
		return
	}

	event := logEntryData{Key: key, Value: value}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "json marshal err" + err.Error(), "data": struct{}{}})
	}
	applyFuture := s.node.raft.Apply(eventJSON, 5*time.Second)

	if err := applyFuture.Error(); err != nil {
		c.JSON(200, gin.H{"code": 1, "error_messaage": "raft apply err" + err.Error(), "data": struct{}{}})
	}

	c.JSON(200, gin.H{
		"code":           0,
		"error_messaage": "",
		"data":           struct{}{},
	})
}

type joinReq struct {
	PeerAddress string `json:"peer_address"`
}

// Join node to cluster
func (s *Server) Join(c *gin.Context) {
	var req joinReq
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "json parse err:" + err.Error(), "data": struct{}{}})
		return
	}

	peerAddress := req.PeerAddress
	if peerAddress == "" {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "peerAddress is empty", "data": struct{}{}})
		return
	}

	addPeerFuture := s.node.raft.AddVoter(raft.ServerID(peerAddress), raft.ServerAddress(peerAddress), 0, time.Second)
	if err := addPeerFuture.Error(); err != nil {

		c.JSON(200, gin.H{"code": 1, "error_messaage": "fail joining peer to raft: " + err.Error(), "data": struct{}{}})
		return
	}

	c.JSON(200, gin.H{
		"code":           0,
		"error_messaage": "",
		"data":           struct{}{},
	})
}
