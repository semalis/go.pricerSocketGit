package TCP

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Client struct {
	sync.RWMutex

	tcp net.Conn

	r *bufio.Reader
	w *bufio.Writer

	outMsg chan []byte

	onMessage func(c *Client, msg []byte) // Коллбек входящий сообщений
	onClose   func(c *Client)             // Коллбек дисконекта
	onConnect func(c *Client)

	cancel context.CancelFunc
	ctx    context.Context

	closed bool
}

func NewClient(conn net.Conn) *Client {
	client := &Client{
		tcp:    conn,
		r:      bufio.NewReader(conn),
		w:      bufio.NewWriter(conn),
		outMsg: make(chan []byte),
	}

	return client
}

// Отвал конекта
func (c *Client) Close() {
	c.Lock()

	if c.closed {
		c.Unlock()
		return
	}

	c.closed = true

	// Чтоб убились горутины
	c.cancel()

	if err := c.tcp.Close(); err != nil {
		log.Printf("Error: Client: Close: %s", err.Error())
	}

	close(c.outMsg)

	c.Unlock()

	//log.Printf("Client [%s]: Disconnect", c.tcp.RemoteAddr())

	c.onClose(c)
}

// Настройка каналов чтения/записи
func (c *Client) SetupChannels() {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	go c.writer()
	go c.reader()
}

// Писатель
func (c *Client) writer() {
	defer func() {
		c.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return

		case msg := <-c.outMsg:
			if _, err := c.w.Write(msg); err != nil {
				fmt.Printf("Client [%s]: Writing error: %s", c.tcp.RemoteAddr(), err.Error())
				return
			}

			if err := c.w.Flush(); err != nil {
				fmt.Printf("Client [%s]: Writing flush error: %s", c.tcp.RemoteAddr(), err.Error())
				return
			}
		}
	}
}

// Читатель
func (c *Client) reader() {
	defer func() {
		c.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := c.r.ReadBytes(Delimiter)

			if err != nil {
				if err != io.EOF {
					fmt.Printf("Client [%s]: Reading error: %s", c.tcp.RemoteAddr(), err.Error())
				}

				return
			}

			c.onMessage(c, msg[:len(msg)-1])
		}
	}
}

// Отправка сообщения
func (c *Client) Send(msg []byte, err error) {
	c.RLock()

	if !c.closed {
		//log.Printf("Send to client [%s]: %s", c.tcp.RemoteAddr(), string(msg))

		msg = append(msg, Delimiter)

		c.outMsg <- msg
	}

	c.RUnlock()
}

// Колбек при входящем сообщении
func (c *Client) OnMessage(f func(client *Client, msg []byte)) {
	c.onMessage = f
}

// Колбек при разрыве соединения
func (c *Client) OnClose(f func(client *Client)) {
	c.onClose = f
}

func (c *Client) OnConnect(f func(client *Client)) {
	c.onConnect = f
}

func (c *Client) GetAddr() string {
	return c.tcp.RemoteAddr().String()
}
