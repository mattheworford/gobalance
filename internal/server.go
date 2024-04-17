package internal

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

type Server struct {
	config      ServerConfig
	isAlive     bool
	statusMutex sync.RWMutex
	proxy       *httputil.ReverseProxy
	logger      *log.Logger
}

type ServerConfig struct {
	URL       string
	Timeout   time.Duration
	MaxConns  int
	TLSConfig *tls.Config
}

func (s *Server) CheckHealth() bool {
	s.statusMutex.RLock()
	defer s.statusMutex.RUnlock()

	if !s.isAlive {
		return false
	}

	if !s.externalHealthCheck() {
		s.logger.Println("External health check failed")
		return false
	}

	return true
}

func (s *Server) externalHealthCheck() bool {
	client := &http.Client{Timeout: s.config.Timeout}
	resp, err := client.Get(s.config.URL)
	if err != nil {
		s.logger.Printf("Failed to connect to the server: %v", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func NewServer(config ServerConfig, logger *log.Logger) *Server {
	return &Server{
		config:  config,
		isAlive: true,
		proxy: &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = config.URL
			},
		},
		logger: logger,
	}
}
