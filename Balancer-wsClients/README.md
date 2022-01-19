# Balancer websocket clients
### rest repiter input and redirecting to any client

Запросы:  
- Подключен repiter, можно отправлять запросы аналогичные как для **rest-repiter**.
- Отправка из очереди осуществляется рандомному клиенту из списка.
- Отправка запросов определенному клиенту

Список подключенных клиентов  
```(json)
method: POST
req_url: http://balancer-server-host.ru:1331/client/rpc
body: {"endpoint": "GetListClients",
       "service": "go.tracker.svc.hubServices"}
response: [
              "193.126.51.168",
              "195.16.78.61"
          ]
``` 


Получить
```sh
root@772581-ce60489:~/balancer-repiter# chmod 777 ./server-balancer-repiter 
root@772581-ce60489:~/balancer-repiter# ./server-balancer-repiter 
config path: /root/balancer-repiter/config.json
2022-01-16 23:47:36 | FUNC:Daemon | UNIT: nil | TEXT: █║ Starting HTTP Listener ▌│║ on port: 1331

2022-01-16 23:47:36 | FUNC:Front Http Server | UNIT: nil | TEXT: Server is started
2022-01-16 23:47:36 | FUNC:Daemon | UNIT: waitForSignal | TEXT: Wait got signal: exiting
2022-01-16 23:47:46 | FUNC:WsHubServices | UNIT: ServeWs | TEXT: подключение к HUB WS client 【193.106.51.168】
2022-01-16 23:47:46 | FUNC:WsHubServices | UNIT: ServeWs | TEXT: подключился WS client 【193.106.51.168】
2022-01-16 23:47:46 | FUNC:WS_Client | UNIT: readPump | TEXT: init
2022-01-16 23:47:46 | FUNC:WS_Client | UNIT: readPump | TEXT: wait new message from ws client
2022-01-16 23:47:46 | FUNC:WsHubServices | UNIT: register | TEXT: 【193.106.51.168】
[LOGIX]:-Client.writePump->init
[LOGIX]:-Client.writePump->for[circle]
```
`
{
	"service": "go.tracker.svc.capproc",
    "endpoint": "Capproc.GetCapsStates",
    "request": {
		"offerId": 90
	}
}
`

CGO_ENABLED=0 GOOS=linux go build -gcflags "all=-N -l" -o balancer-repiter
