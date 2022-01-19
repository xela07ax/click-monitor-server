package wsHandler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"time"
)

type ConnecterWsServer struct {
	url          url.URL
	client       *Client
	Done         bool
	balancerHost string
	exec         func([]byte) []byte
	Loger        chan<- [4]string
}

func NewWsDaemon(balancerHost string, exec func([]byte) []byte, loger chan<- [4]string) *ConnecterWsServer {
	// Прописываем пути
	cws := &ConnecterWsServer{
		url:          url.URL{Scheme: "ws", Host: balancerHost, Path: "/client/ws"},
		Done:         false,
		balancerHost: balancerHost,
		exec:         exec,
		Loger:        loger,
	}
	cws.RunServe()
	return cws
}

func (c *ConnecterWsServer) RunServe() {
	go func() {
		err := c.tryConnect()
		if err != nil {
			c.Loger <- [4]string{"ConnecterWsServer", "RunServe", fmt.Sprintf("ошибка: %s | server:%s | sleep: 30 ", err, c.url.String()), "ERROR"}
			time.Sleep(30 * time.Second)
			c.RunServe()
		}
	}()
}

func (c *ConnecterWsServer) tryConnect() error {
	c.Loger <- [4]string{"wsHandler", "NewWsConnect", fmt.Sprintf("начинаем подключение к %s", c.url.String())}
	connws, _, err := websocket.DefaultDialer.Dial(c.url.String(), nil)
	if err != nil {
		return fmt.Errorf("подключение к серверу не удалось |err:%v", err)
	}
	c.Loger <- [4]string{"wsHandler", "NewWsConnect", "подключение успешно"}

	c.client = &Client{hub: c, exec: c.exec, conn: connws, send: make(chan []byte, 256), Loger: c.Loger}

	go c.client.writePump()
	go c.client.readPump()

	return nil
}
