package reqestWrkHub

import (
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"time"
)

// Client is a middleman between the hub.
type Client struct {
	hub        *Hub
	sender     *model.Sender
	pauseStart time.Time
	pauseEnd   time.Time
}

type Hub struct {
	cfg        *model.Config
	clients    map[string]*Client
	paused     map[string]*Client
	validate   func(string) bool
	Input      chan model.IpqsPrepareData
	register   chan *Client
	unregister chan *Client
	pause      chan *Client
	LogIpqs    chan model.IPQSRow
	LogBad     chan [2]string
	Loger      chan<- [4]string
}

func NewHub(cfg *model.Config, validate func(string) bool, loger chan<- [4]string) *Hub {
	hub := &Hub{
		cfg:        cfg,
		Input:      make(chan model.IpqsPrepareData, cfg.CashLenth),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		pause:      make(chan *Client),
		clients:    make(map[string]*Client),
		paused:     make(map[string]*Client),
		validate:   validate,
		LogIpqs:    make(chan model.IPQSRow, 500),
		LogBad:     make(chan [2]string, 500),
		Loger:      loger,
	}
	go hub.RunHub()
	go hub.RunWakeUp()
	// добавим доступных в конфиге клиентов
	for _, sender := range cfg.Senders {
		hub.AddClient(sender)
	}

	loger <- [4]string{"NewHub", "хаб иницаализирован", "хаб отправителей готов к работе"}
	return hub
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client.sender.Name]; ok {
		h.Loger <- [4]string{"Hub.DaemonHub.unregister", "удаляем клиента из списка активных", fmt.Sprintf("clientName:%s", client.sender.Name), "INFO"}
		delete(h.clients, client.sender.Name)
	}
}

func (h *Hub) RunWakeUp() {
	// будем раз в 30 секунд проверять, может кого нибудь из паузы вытащим
	ticker := time.Tick(30 * time.Second)
	for {
		<-ticker
		for k, v := range h.paused {
			dur := v.pauseEnd.Sub(time.Now())
			if dur < 1 {
				h.Loger <- [4]string{"Hub.PauserTicker.wakeup", "клиент разблокирован", fmt.Sprintf("pauseStart:%v|pauseEnd:%v|name:%s|lastTime:%d|sleep:%d", v.pauseStart, v.pauseEnd, k, dur*time.Minute, v.sender.Sleep), "INFO"}
				v.sender.ErrDetect = false
				h.register <- v
			}
		}
	}
}

func (h *Hub) RunHub() {
	for {
		select {
		case client := <-h.register:
			if client.sender.ErrDetect {
				// не регестрировать клиента который имеет ошибку
				h.Loger <- [4]string{"Hub.DaemonHub.register", "не можем зарегестрировать клиента который имеет ошибку", fmt.Sprintf("clientName:%s| отправка на паузу", client.sender.Name), "WARNING"}
				go func() { h.pause <- client }()
				continue
			}
			if _, ok := h.clients[client.sender.Name]; !ok {
				// клиента нет в списке активных, добавляем
				h.Loger <- [4]string{"Hub.DaemonHub.register", "добавляем клиента в список активных", fmt.Sprintf("clientName:%s", client.sender.Name), "INFO"}
				h.playClient(client)
				h.clients[client.sender.Name] = client
				continue
			}
			h.Loger <- [4]string{"Hub.DaemonHub.register", "попытка добавить уже добавленного клиента", "ошибока, операция не может быть выполнена", "ERROR"}
		case client := <-h.unregister:
			h.unregisterClient(client)
		case client := <-h.pause:
			h.unregisterClient(client)
			// надо рассчитать время разблокировки клиента
			client.pauseStart = time.Now()
			client.pauseEnd = client.pauseStart.Add(time.Duration(client.sender.Sleep) * time.Minute)
			h.Loger <- [4]string{"Hub.DaemonHub.pause", "помещен на паузу", fmt.Sprintf("clientName:%s|pauseMinute:%d|startPause:%v|endPause:%v", client.sender.Name, client.sender.Sleep, client.pauseStart, client.pauseEnd), "ERROR"}
			h.paused[client.sender.Name] = client
		}
	}
}
