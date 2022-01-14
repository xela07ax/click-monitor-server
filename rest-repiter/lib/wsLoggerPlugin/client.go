package wsLoggerPlugin

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xela07ax/toolsXela/ttp"
	"io/ioutil"
	"log"
	"net/http"
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
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *WsLogger

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	Loger chan<- [4]string
}

func (c *Client) recoveryReader() {
	if rMsg := recover(); rMsg != nil { // Если была паника, будем отвечать ошибкой в сентри и лог
		err := fmt.Errorf("recoveryReader[wsLoggerPlugin]info[детектирована паника]err:%v", rMsg)
		c.Loger <- [4]string{"Handler", "json.Unmarshal", fmt.Sprintf("err:%v", err), "PANIC"}
	}
}
// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	c.hub.Loger <- [4]string{"WS_Client", "readPump", "init"}
	defer c.recoveryReader()
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	c.hub.Loger <- [4]string{"WS_Client", "readPump", "wait new message from ws client"}
	for {
		_, message, err := c.conn.ReadMessage()
		c.hub.Loger <- [4]string{"WS_Client", "readPump", string(message)}
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.send <- []byte(fmt.Sprintf("█║ (•̪●)=︻╦̵̵̿╤──:%s", message))
		err, resp := c.hub.Interpretator(message)
		if err != nil {
			c.send <- []byte(fmt.Sprintf("█║ ¯\\_(ツ)_/¯ err:%s", err))
			continue
		}
		if len(resp) > 0 {
			c.send <- resp
		}
		//c.hub.broadcast <- message
	}
}

func (c *Client) recoveryWriter() {
	if rMsg := recover(); rMsg != nil { // Если была паника, будем отвечать ошибкой в сентри и лог
		err := fmt.Errorf("writePump[wsLoggerPlugin]info[детектирована паника]err:%v", rMsg)
		c.Loger <- [4]string{"Handler", "json.Unmarshal", fmt.Sprintf("err:%v", err), "PANIC"}
	}
}
// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	defer c.recoveryWriter()
	Logx("-Client.writePump->init")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		Logx("-Client.writePump->defer func[X]")
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		Logx("-Client.writePump->for[circle]")
		select {
		case message, ok := <-c.send:
			Logx(fmt.Sprintf("-Client.writePump->for-message[%s]",message))
			log.Printf("msg:%s",message)
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
			Logx(fmt.Sprintf("-Client.writePump->for-message[%s]-ok",message))
		case <-ticker.C:
			Logx(fmt.Sprintf("-Client.writePump->for-ticker.C-SetWriteDeadline"))
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func readBodySimple(w http.ResponseWriter, r *http.Request) []byte {
	fmt.Println("Пришла команда по HTTP 1.1")
	b, err := ioutil.ReadAll(r.Body) // Считывание тело, ожидаем адрес сервера "localhost:5578"
	if err != nil {
		ertx := fmt.Sprintf("COM:Ошибка чтения тела: %s | ERTX:can't read body", err)
		fmt.Println(ertx)
		http.Error(w, ertx, http.StatusConflict) // 409
		return []byte{}
	}
	return b
}
type Notify struct {
	FuncName string
	Text string
	Status int
	Show bool
	UpdNum int
}

func resp (w http.ResponseWriter, r *http.Request, funcName string,text string, status int, show bool) {
	Notify := Notify {
		FuncName: funcName,
		Text:     text,
		Status:   status,
		Show:     show,
	}
	if err := ttp.Httpjson(w, r, Notify); err != nil {
		log.Fatalf("Критическая ошибка, не удалось отправить сообщение в UI: %s| %v", err,Notify)
	}

}


