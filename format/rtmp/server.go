package rtmp

import (
	"bufio"
	"net"
	"time"
)

const (
	EventConnConnected     = 1
	EventHandshakeFailed   = 2
	EventConnDisconnected  = 4
	EventConnConnectFailed = 5
)

var EventString = map[int]string{
	EventConnConnected:     "Connected",
	EventConnConnectFailed: "ConnectFailed",
	EventHandshakeFailed:   "HandshakeFailed",
	EventConnDisconnected:  "ConnDisconnected",
}

type Server struct {
	OnNewConn  func(c *Conn)
	HandleConn func(c *Conn, nc net.Conn)

	HandshakeTimeout time.Duration

	LogEvent func(c *Conn, nc net.Conn, e int)
}

func NewServer() *Server {
	return &Server{
		HandshakeTimeout: time.Second * 10,
	}
}

type bufReadWriter struct {
	*bufio.Reader
	*bufio.Writer
}

var BufioSize = 4096

func (s *Server) handleAcceptConn(nc net.Conn) {
	rw := &bufReadWriter{
		Reader: bufio.NewReaderSize(nc, BufioSize),
		Writer: bufio.NewWriterSize(nc, BufioSize),
	}
	c := NewConn(rw)
	c.isserver = true

	if fn := s.OnNewConn; fn != nil {
		fn(c)
	}

	if s.LogEvent != nil {
		s.LogEvent(c, nc, EventConnConnected)
	}

	nc.SetDeadline(time.Now().Add(time.Second * 15))
	if err := c.Prepare(StageGotPublishOrPlayCommand, 0); err != nil {
		if s.LogEvent != nil {
			s.LogEvent(c, nc, EventHandshakeFailed)
		}
		nc.Close()
		return
	}
	nc.SetDeadline(time.Time{})

	s.HandleConn(c, nc)
}

func (s *Server) Serve(lis net.Listener) (err error) {
	for {
		var nc net.Conn
		if nc, err = lis.Accept(); err != nil {
			return
		}
		go s.handleAcceptConn(nc)
	}
}
