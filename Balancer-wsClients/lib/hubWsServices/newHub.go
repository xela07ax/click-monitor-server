package hubWsServices

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/rest-repiter/model"
	"net"
	"net/http"
	"time"
)

func Logx(txt string) {
	fmt.Printf("[LOGIX]:%s\n", txt)
}
func NewWsConnector(execInputMsg func([]byte) []byte, loger chan<- [4]string) *WsHubServices {
	return &WsHubServices{
		execInputMsg:  execInputMsg,
		broadcast:     make(chan []byte),
		Input:         make(chan []byte, 100),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]int),
		//bufferClients: make(chan *Client, 100),
		Loger:         loger,
	}
}

func (hub *WsHubServices) ServeWs(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	hub.Loger <- [4]string{"WsHubServices", "ServeWs", fmt.Sprintf("подключение к HUB WS client 【%s】", ip)}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.Loger <- [4]string{"WsHubServices", "ServeWs", fmt.Sprintf("[upgrader.Upgrade]err:%v", err), "ERROR"}
	}
	client := &Client{hub: hub, ip: ip, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client
	hub.Loger <- [4]string{"WsHubServices", "ServeWs", fmt.Sprintf("подключился WS client 【%s】", ip)}
	// Разрешить сбор памяти, на которую ссылается вызывающий абонент, выполнив всю работу в
	// новых goroutines.
	go client.writePump()
	go client.readPump()
}

func (hub *WsHubServices) SentMsgFromAnyClient(reqData []byte) []byte {
	for k, _ := range hub.clients {
		hub.Loger <- [4]string{"WsHubServices", "SentMsgFromAnyClient", fmt.Sprintf("Отправка сообщения RPC клиенту: [%s]", k.ip)}

		req, _ := json.Marshal(model.RpcModel{
			Micro:        model.Micro{
				Service:  "go.tracker.svc.repiter",
				Endpoint: "",
			},
			MicroRequest: model.MicroRequest{Request: string(reqData)},
		})
		return hub.respWaitClient(k, req)
	}
	return  []byte("[BALANCER]err: ни одного клиента не подключено")
}


func (hub *WsHubServices) respWaitClient(client *Client, reqData []byte) []byte {
	// приготовимся к ответу и запросу
	respChan := make(chan []byte)
	client.waitResp = respChan
	defer func() {
		client.waitResp = nil
	}()

	client.send <- reqData

	select {
	case respClient := <-respChan:
		return respClient
	case <-time.After(30 * time.Second):
		return []byte(fmt.Sprintf("[BALANCER]err: таймаут ожидания ответа от клиента: 30 сек"))
	}
}
func (hub *WsHubServices) SentMsgFromClient(reqData []byte) []byte {
	req := new(model.MicroExec)
	err := json.Unmarshal(reqData, req)
	if err != nil {
		return []byte(fmt.Sprintf("не удалось распознать свойства запроса| err:%v", err))
	}
	client, err := hub.getClient(req.Param)
	if err != nil {
		return []byte(fmt.Sprintf("не удалось отправить запрос| err:%v", err))
	}
	return hub.respWaitClient(client, []byte(req.Request))
}

// если брать клиентов по ip то получается, что нельзя запускать больше одного экземпляра на хосте
func (hub *WsHubServices) getClient(ipClient string) (client *Client, err error) {
	for client, _ := range hub.clients {
		if client.ip == ipClient {
			return client, nil
		}
	}
	return nil, fmt.Errorf("ошибка: клиент [%s] не найден или отключен", ipClient)
}
func (hub *WsHubServices) GetListClients([]byte) (clients []byte) {
	clients, err := json.Marshal(hub.getListClients())
	if err != nil {
		panic(fmt.Errorf("%v|func:GetListClients", clients))
	}
	hub.Loger <- [4]string{"WsHubServices", "GetListClients", fmt.Sprintf("принят запрос на выдачу своих клиентов по ip【%s】", clients)}
	return
}
func (hub *WsHubServices) getListClients() (clients []string) {
	for k, _ := range hub.clients {
		clients = append(clients, k.ip)
	}
	return
}
