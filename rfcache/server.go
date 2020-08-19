package rfcache

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
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
		c.JSON(400, gin.H{"code": 1, "error_messaage": "key is empty", "data": []interface{}{}})
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
		c.JSON(400, gin.H{"code": 1, "error_messaage": "json parse err:" + err.Error(), "data": []interface{}{}})
		return
	}

	key := req.Key
	if key == "" {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "key is empty", "data": []interface{}{}})
		return
	}
	value := req.Value
	if value == "" {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "value is empty", "data": []interface{}{}})
		return
	}

	event := logEntryData{Key: key, Value: value}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		c.JSON(400, gin.H{"code": 1, "error_messaage": "json marshal err" + err.Error(), "data": []interface{}{}})
	}
	applyFuture := s.node.raft.Apply(eventJSON, 5*time.Second)

	if err := applyFuture.Error(); err != nil {
		c.JSON(200, gin.H{"code": 1, "error_messaage": "raft apply err" + err.Error(), "data": []interface{}{}})
	}

	c.JSON(200, gin.H{
		"code":           0,
		"error_messaage": "",
		"data":           []interface{}{},
	})
}
