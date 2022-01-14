package wsLoggerPlugin

import "fmt"

type WsLogger struct {
	// Зарегистрированный клиент.
	clients map[*Client]bool
	Input chan []byte
	Interpretator func([]byte)(error, []byte)
	// Входящие сообщения от клиентов.
	broadcast chan []byte

	// Регистрируйте запросы от клиентов.
	register chan *Client

	// Отмените регистрацию запросов от клиентов.
	unregister chan *Client

	Loger chan <- [4]string
}



func (h *WsLogger) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
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
			// Входящие сообщения от клиентов.
			// Собственно работать будем тут
			for client := range h.clients {
				select {
				case client.send <- message:
					fmt.Printf("-hub.run->select[h.broadcast]for-select[%s]\n",message)
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
