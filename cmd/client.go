package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type client struct {
	serverAddress string
	clientAddress string
	conn          net.Conn
	in            io.Reader
	quitCh        chan struct{}
}

func newClient(address string, in io.Reader) *client {
	return &client{
		serverAddress: address,
		in:            in,
		quitCh:        make(chan struct{}),
	}
}

func (c *client) connectToChat() error {
	conn, err := net.Dial("tcp", c.serverAddress)
	if err != nil {
		return fmt.Errorf("could not connect to chat: %v", err)
	}
	defer conn.Close()
	c.conn = conn
	c.clientAddress = conn.LocalAddr().String()

	var wg sync.WaitGroup
	wg.Add(2)
	go c.getUpdatesFromServer(&wg)
	go c.getInput(&wg)

	wg.Wait()
	return nil
}

func (c *client) getInput(wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(c.in)
	for scanner.Scan() {
		_, err := c.conn.Write([]byte(scanner.Text()))
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

// getUpdatesFromServer обрабатывает сообщения от сервера
func (c *client) getUpdatesFromServer(wg *sync.WaitGroup) {
	defer wg.Done()
	buf := make([]byte, 1024)
	for {
		select {
		case <-c.quitCh:
			return
		default:
			c.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			n, err := c.conn.Read(buf)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
				} else {
					return
				}
			}
			fmt.Print(string(buf[:n]))
		}
	}
}
