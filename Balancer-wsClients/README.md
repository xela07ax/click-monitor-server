# Balancer websocket clients
### rest repiter input and redirecting to any client

Запросы:  
- Подключен repiter, можно отправлять запросы аналогичные как для **rest-repiter**.
- Отправка из очереди осуществляется рандомному клиенту из списка.
```(json)
{
    "service": "go.tracker.svc.repiter",
    "endpoint": "",
    "request": "{\"Time\":\"2022-01-15T00:24:13.1951894+03:00\",\"Send\":true,\"RedirectUrl\":\"http://trackerhqu.com/15b1/602/39b/86a4937a-5dac-53a5-9f4f-5d137d5c030d\",\"Params\":\"allow_public_access_points=true\\u0026fast=false\\u0026lighter_penalties=true\\u0026mobile=false\\u0026strictness=1\",\"Method\":\"GET\",\"Body\":\"\",\"Type\":\"\"}"
}
```
- Отправка запросов определенному клиенту
- Список подключенных клиентов  
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


Компиляция
```sh
CGO_ENABLED=0 GOOS=linux go build -gcflags "all=-N -l" -o balancer-repiter
```

Немного логов
```sh
root@772581-ce60489:~/balancer-repiter# chmod 777 ./server-balancer-repiter 
root@772581-ce60489:~/balancer-repiter# ./server-balancer-repiter 
config path: /root/balancer-repiter/config.json
2022-01-16 23:47:36 | FUNC:Daemon | UNIT: nil | TEXT: █║ Starting HTTP Listener ▌│║ on port: 1331

2022-01-16 23:47:36 | FUNC:Front Http Server | UNIT: nil | TEXT: Server is started
2022-01-16 23:47:36 | FUNC:Daemon | UNIT: waitForSignal | TEXT: Wait got signal: exiting
2022-01-16 23:47:46 | FUNC:WsHubServices | UNIT: ServeWs | TEXT: подключение к HUB WS client 【193.16.51.168】
2022-01-16 23:47:46 | FUNC:WsHubServices | UNIT: ServeWs | TEXT: подключился WS client 【193.16.51.168】
2022-01-16 23:47:46 | FUNC:WS_Client | UNIT: readPump | TEXT: init
2022-01-16 23:47:46 | FUNC:WS_Client | UNIT: readPump | TEXT: wait new message from ws client
2022-01-16 23:47:46 | FUNC:WsHubServices | UNIT: register | TEXT: 【193.16.51.168】
[LOGIX]:-Client.writePump->init
[LOGIX]:-Client.writePump->for[circle]
```
