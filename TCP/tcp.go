package TCP

import (
	"log"
	"net"
)

const Delimiter = byte('\n')

type Server struct {
	address   string
	onMessage func(client *Client, msg []byte)
	onClose   func(client *Client)
	onConnect func(client *Client)

	listener net.Listener
}

func NewServer(address string) *Server {
	server := &Server{
		address: address,
	}

	go server.listen()

	return server
}

func (s *Server) OnMessage(cb func(client *Client, msg []byte)) {
	s.onMessage = cb
}

func (s *Server) OnClose(cb func(client *Client)) {
	s.onClose = cb
}

func (s *Server) OnConnect(cb func(client *Client)) {
	s.onConnect = cb
}

func (s *Server) listen() {
	listener, err := net.Listen("tcp", s.address)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("TCP Error: %s", err.Error())
		}
	}()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("TCP Error: %s", err.Error())
		}

		s.handler(conn)
	}
}

func (s *Server) handler(conn net.Conn) {
	//log.Println("New connection:", conn.RemoteAddr())

	client := NewClient(conn)

	client.OnConnect(s.onConnect)
	client.OnMessage(s.onMessage)
	client.OnClose(s.onClose)

	client.SetupChannels()

	client.onConnect(client)
}
