package hubWsServices

import "fmt"

type WsHubServices struct {
	execInputMsg func([]byte)[]byte
	clients map[*Client]int // Зарегистрированный клиент.
	Input   chan []byte
	broadcast chan []byte // отправлять сообщения для всех клиентов.
	register chan *Client // Регистрируйте запросы от клиентов.
	unregister chan *Client // Отмените регистрацию запросов от клиентов.
	//bufferClients chan *Client
	Loger         chan<- [4]string
}

func (h *WsHubServices) Run() {
	for {
		select {
		case client := <-h.register:
			h.Loger <- [4]string{"WsHubServices", "register", fmt.Sprintf("【%s】", client.ip)}
			h.clients[client] = 1
		case client := <-h.unregister:
			h.Loger <- [4]string{"WsHubServices", "unregister", fmt.Sprintf("【%s】", client.ip)}
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				//for {
				//	//tmpClient := <-h.bufferClients
				//	listTmpClients := make([]*Client, 0, 4)
				//	if tmpClient != client {
				//		listTmpClients = append(listTmpClients, tmpClient)
				//	}
				//	for _, tmpClient := range listTmpClients {
				//		h.bufferClients <- tmpClient
				//	}
				//}
				close(client.send)
			}
		case message := <-h.Input:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
					fmt.Printf("-hub.run->select[h.broadcast]for-select[%s]\n", message)
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
