# Repiter - Сlient from balancer
_Перенаправлять запросы (можно и др. команды) с любого хоста, подключаемся по web socket к серверу._

Запуск  
```sh
./client-repiter BALANCER_HOST:PORT
client-repiter.exe 123.245.46.78:1331
```

Информация
- модуль для Balancer: **wsClient-rest-repiter** - отправляем запросы туда, отправятся отсюда
- если сервер не доступен, то будем опрашивать для подключения
- размер сообщения ограничен небольшой страничкой, иначе ответ не вернется

Немного логов в хату, там есть запрос:
```
2022-01-17 00:28:10 | FUNC:main | UNIT: args | TIP:INFO |TEXT: 【параметры хоста сервера переданы: 195.212.10.101:1331】
2022-01-17 00:28:10 | FUNC:main | UNIT: args | TIP:INFO |TEXT: 【подключен rpc сервис: go.tracker.svc.repiter】
2022-01-17 00:28:10 | FUNC:wsHandler | UNIT: NewWsConnect | TIP: |TEXT: 【начинаем подключение к ws://195.212.10.101:1331/client/ws】
2022-01-17 00:28:10 | FUNC:wsHandler | UNIT: NewWsConnect | TIP: |TEXT: 【подключение успешно】
2022-01-17 00:29:00 | FUNC:wsHandler | UNIT: Client.readPump | TIP:REQUEST |TEXT: 【{"service":"go.tracker.svc.repiter","request":"{\"Time\":\"2022-01-15T03:23:26.118882+03:00\",\"Send\":true,\"RedirectUrl\":\"https://api.whatismybrowser.com/api/v2/user_agent_parse\",\"Method\":\"POST\",\"Body\":\"{}\"}"}】
2022-01-17 00:29:00 | FUNC:ⓇⓅⒸ | UNIT: ⚡𝓻𝓮𝓺𝓾𝓮𝓼𝓽⚡ | TIP:HTTP_READ |TEXT: 【🅱🅾🅳🆈【{"service":"go.tracker.svc.repiter","request":"{\"Time\":\"2022-01-15T03:23:21188882+03:00\",\"Send\":true,\"RedirectUrl\":\"https://api.whatismybrowser.com/api/v2/user_agent_parse\",\"Method\":\"POST\",\"Body\":\"{}\"}"}】】
2022-01-17 00:29:00 | FUNC:Sent | UNIT: POST | TIP: |TEXT: 【{2022-01-15 03:23:26.1188882 +0300 MSK true https://api.whatismybrowser.com/api/v2/user_agent_parse  POST { }】
2022/01/17 00:29:01 msg:{"RespStatus":"401 Unauthorized","RespCode":401,"RespBody":"{\"result\": {\"code\": \"error\", \"message_code\": \"missing_api_authentication\", \"message\": \"No X-API-KEY header was provided in the request\"}}"}
2022-01-17 00:29:01 | FUNC:Sent | UNIT: https://api.whatismybrowser.com/api/v2/user_agent_parse | TIP:RESPONSE |TEXT: 【{401 Unauthorized 401 {"result": {"code": "error, "message_code": "missing_api_authentication", "message": "No X-API-KEY header was provided in the request"}}}】
2022-01-17 00:29:01 | FUNC:wsHandler | UNIT: Client.readPump | TIP:RESPONSE |TEXT: 【{"RespStatus":"401 Unauthorized","RespCode":401,"RespBody":"{\"result\": {\"code\":\"error\", \"message_code\": \"missing_api_authentication\", \"message\": \"No X-API-KEY header was provided in the request\"}}"}】
```
Компиляция
```sh
CGO_ENABLED=0 GOOS=linux go build -gcflags "all=-N -l" -o client-repiter
CGO_ENABLED=0 GOOS=windows go build -gcflags "all=-N -l" -o client-repiter.exe
```