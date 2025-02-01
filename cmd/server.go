package main

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type connectedClient struct {
	conn net.Conn
	name string
	ch   chan<- string
}

type server struct {
	listenAddr string
	ln         net.Listener
	peerMap    map[connectedClient]bool
	entering   chan connectedClient
	leaving    chan connectedClient
	messages   chan string
	quitCh     chan struct{}
}

func newServer(listenAddr string) *server {
	s := &server{
		listenAddr: listenAddr,
		peerMap:    make(map[connectedClient]bool),
		entering:   make(chan connectedClient),
		leaving:    make(chan connectedClient),
		messages:   make(chan string),
		quitCh:     make(chan struct{}),
	}
	return s
}

func (s *server) start() {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	s.ln = ln

	var wg sync.WaitGroup
	wg.Add(2)
	go s.acceptLoop(&wg)
	go s.broadcaster(&wg)
	wg.Wait()

	close(s.messages)
	close(s.entering)
	close(s.leaving)
	s.ln.Close()
}

func (s *server) acceptLoop(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-s.quitCh:
			return
		default:
		}

		conn, err := s.ln.Accept()
		if err != nil {
			continue
		}
		go s.handleConn(conn)
	}
}

// broadcaster рассылает входящие сообщения всем клиентам
// следит за подключением и отключением клиентов
func (s *server) broadcaster(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-s.quitCh:
			return
		default:
		}

		select {
		case m := <-s.messages:
			for c := range s.peerMap {
				c.ch <- m
			}
		case e := <-s.entering:
			s.peerMap[e] = true
		case l := <-s.leaving:
			delete(s.peerMap, l)
		}
	}
}

// handleConn обрабатывает входящие сообщения от клиента
func (s *server) handleConn(conn net.Conn) {
	defer conn.Close()
	ch := make(chan string)
	defer close(ch)

	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	cli := connectedClient{conn, who, ch}

	ch <- "You are " + who
	s.messages <- who + " has arrived"
	s.entering <- cli

	buf := make([]byte, 1024)
	for {
		select {
		case <-s.quitCh:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			n, err := conn.Read(buf)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue
				}
				s.leaving <- cli
				s.messages <- who + " has left"
				return
			}
			s.messages <- who + ": " + string(buf[:n])
		}
	}
}

// clientWriter отправляет сообщения текущему клиенту
func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		_, err := conn.Write([]byte(msg + "\n"))
		if err != nil {
			return
		}
	}
}

func main() {
	s := newServer("localhost:8000")
	s.start()
}
