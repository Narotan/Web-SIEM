package server

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"nosql_db/internal/api"
	"nosql_db/internal/handlers"
	"time"
)

type TCPServer struct {
	Address       string
	Timeout       int
	MaxConnection int
}

func New(address string) *TCPServer {
	return &TCPServer{
		Address:       address,
		Timeout:       60,
		MaxConnection: 100,
	}
}

func (s *TCPServer) Run() error {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("server running on %s", s.Address)

	maxOpenConntecion := make(chan any, s.MaxConnection)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("conn error: %v", err)
			continue
		}

		maxOpenConntecion <- struct{}{}

		go func() {
			s.handleConnection(conn)

			<-maxOpenConntecion
		}()
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	timeoutDuration := time.Duration(s.Timeout) * time.Second

	_ = conn.SetDeadline(time.Now().Add(timeoutDuration))

	clientAddr := conn.RemoteAddr().String()
	log.Printf("client connected: %s", clientAddr)

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		_ = conn.SetDeadline(time.Now().Add(timeoutDuration))

		var req api.Request
		err := decoder.Decode(&req)
		if err != nil {
			if err == io.EOF {
				log.Printf("client disconnected: %s", clientAddr)
			} else {
				log.Printf("decode error from %s: %v", clientAddr, err)
			}
			return
		}

		resp := handlers.HandleRequest(req)

		_ = conn.SetDeadline(time.Now().Add(timeoutDuration))

		if err := encoder.Encode(resp); err != nil {
			log.Printf("encode error to %s: %v", clientAddr, err)
			return
		}
	}
}
