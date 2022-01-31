package reqestWrkHub

import (
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"time"
)

func (h *Hub) AddClient(sender *model.Sender) {
	client := &Client{hub: h, sender: sender}
	h.register <- client
}

func (h *Hub) playClient(client *Client) {
	go client.scanPump()
}

func (c *Client) scanPump() {
	for {
		message := <-c.hub.Input
		result, err := c.CallThirdParty(message.Ip, message.UserAgent)

		var ertx string
		if err != nil {
			ertx = fmt.Sprintf("ошибка HTTP запроса:%v", err)
		}
		if c.hub.validate(result.RespBody) {
			ertx = fmt.Sprintf("ответ невалидный:%v", result.RespBody)
		}
		if ertx != "" {
			c.hub.Loger <- [4]string{"Client.errResponse", "поставим клиента на паузу", fmt.Sprintf("name:%s|pauseMinuts:%d", c.sender.Name, c.sender.Sleep)}
			c.hub.LogBad <- [2]string{c.sender.Name, ertx}
			c.hub.pause <- c
			return
		}
		// ошибок нет, ответ валиден
		result.SenderIp = c.sender.Name
		result.Ip = message.Ip
		result.Uag = message.UserAgent
		c.hub.LogIpqs <- result
		// продолжаем сканирование основного канала, через 3 сек
		time.Sleep(3*time.Second)
	}
}