package wsHandler

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub  *ConnecterWsServer
	exec func([]byte) []byte
	// The websocket connection.
	conn *websocket.Conn
	url  url.URL
	// Buffered channel of outbound messages.
	send chan []byte
	Loger chan<- [4]string
}

func (c *Client) recovery() {
	if rMsg := recover(); rMsg != nil { // Если была паника, будем отвечать ошибкой в сентри и лог
		err := fmt.Errorf("recoveryReader[wsHandler]info[детектирована паника]err:%v", rMsg)
		c.Loger <- [4]string{"Handler", "json.Unmarshal", fmt.Sprintf("err:%v", err), "PANIC"}
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.Loger <- [4]string{"wsHandler", "Client.readPump", "остановлено, обрыв коннекта с сервером", "INFO"}
		c.conn.Close()
		c.hub.RunServe()
	}()
	defer c.recovery()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			//if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {}
			c.Loger <- [4]string{"wsHandler", "Client.readPump", err.Error(), "ERROR"}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.Loger <- [4]string{"wsHandler", "Client.readPump", string(message), "REQUEST"}
		resp := c.exec(message)
		c.send <- resp
		c.Loger <- [4]string{"wsHandler", "Client.readPump", string(resp), "RESPONSE"}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	defer c.recovery()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		fmt.Printf("Client.writePump: остановлено\n")
		ticker.Stop()
		//c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			log.Printf("msg:%s", message)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
