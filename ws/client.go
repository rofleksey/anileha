package ws

import (
	"anileha/config"
	"encoding/json"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
)

type Client struct {
	ID           uint
	RoomId       uint
	Conn         *websocket.Conn
	NotifyClose  atomic.Bool
	Closed       atomic.Bool
	sendChan     chan any
	onReceive    func(client *Client, bytes []byte)
	onDisconnect func(client *Client)
	config       *config.Config
	log          *zap.Logger
}

func NewClient(id uint, conn *websocket.Conn, onReceive func(client *Client, bytes []byte),
	onDisconnect func(client *Client), config *config.Config, log *zap.Logger) *Client {
	return &Client{
		ID:           id,
		Conn:         conn,
		config:       config,
		log:          log,
		sendChan:     make(chan any, config.WebSocket.MessageChanBufferSize),
		onReceive:    onReceive,
		onDisconnect: onDisconnect,
	}
}

// Client goroutine to read messages from client
func (c *Client) read() {
	defer func() {
		if c.NotifyClose.Load() {
			c.onDisconnect(c)
		}
		c.Conn.Close()
	}()

	pingTimeout := time.Duration(c.config.WebSocket.PingTimeoutMs) * time.Millisecond

	c.Conn.SetReadLimit(c.config.WebSocket.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pingTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pingTimeout))
		return nil
	})

	for {
		_, bytes, err := c.Conn.ReadMessage()
		if err != nil {
			c.log.Warn("failed to read message from websocket", zap.Error(err))
			break
		}
		c.onReceive(c, bytes)
	}
}

// Client goroutine to write messages to client
func (c *Client) write() {
	pingInterval := time.Duration(c.config.WebSocket.PingIntervalMs) * time.Millisecond
	writeTimeout := time.Duration(c.config.WebSocket.WriteTimeoutMs) * time.Millisecond

	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.sendChan:
			c.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.log.Info("send channel is closed")
				return
			} else {
				jsonBytes, err := json.Marshal(message)
				if err != nil {
					c.log.Warn("failed to write message to websocket", zap.Error(err))
					continue
				}
				err = c.Conn.WriteMessage(1, jsonBytes)
				if err != nil {
					c.log.Warn("failed to write message to websocket", zap.Error(err))
				}
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.log.Warn("websocket ping timeout", zap.Error(err))
				return
			}
		}
	}
}

func (c *Client) Send(msg any) bool {
	select {
	case c.sendChan <- msg:
		return true
	default:
		return false
	}
}

func (c *Client) Close(notifyClose bool) {
	if c.Closed.CompareAndSwap(false, true) {
		c.NotifyClose.Store(notifyClose)
		c.log.Info("closing client")
		close(c.sendChan)
	}
}

func (c *Client) Start() {
	c.log.Info("starting client")
	go c.write()
	go c.read()
}
