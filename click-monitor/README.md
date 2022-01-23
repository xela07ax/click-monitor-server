## Click Monitor
#### Подготовка и отправка запросов с последующим сохранением и кешированием


1) Считывает сообщения из БД о новых кликах
    - мониторинг таблицы на появление новых кликов
2) Подготавливает URL для отправки в сторонний сервис
    - подготавливает, но сам не отправляет, так как его IP не предназначен для таких запросов
3) Отправляет на последующую отправку в ***rest-repiter*** который на стороннем сервере ждет rest сообщения 
4) Сортирует полученный результат на ошибки, ок и найденные в кеше для мини Репортинга 
    - отчет в реальном времени в папке report
    - сохранение в OLTP базу для дальнейшей работы с данными
5) Кеширует запросы по ключу IP и сохраняет UserAgent + quality_score data

Конфигурации лежат в папке bin 

Компиляция:  
```sh
cd click-monitor
CGO_ENABLED=0 GOOS=windows go build -gcflags "all=-N -l" -o click-monitor.exe 
CGO_ENABLED=0 GOOS=linux go build -gcflags "all=-N -l" -o click-monitor
```

Далее необходимо сделать его исполняемым:
```sh
chmod +x ./click-monitor
```
Запуск:  
```sh
./click-monitor
or
./click-monitor 1>log1stdout.txt 2>log1stderr.txt
```
или если хотите оставить демона в системе
```sh
nohup ./click-monitor &
ctrl^C
```

Обработка ошибок, смена клиента
```sh
# Очередной запрос с хоста, но ответ уже показывает, что пора менять хост
2022-01-23 23:05:20 | FUNC:GenFilter.Select | UNIT: rows | TIP:INFO |TEXT: 【extract: 4】
2022-01-23 23:05:20 | FUNC:ⓇⓅⒸ | UNIT: http.Post[http://sky.net.kg:1333/rpc] | TIP:RESPONSE |TEXT: 【body:{"RespStatus":"200 OK","RespCode":200,"RespBody":"{\"success\":false,\"message\":\"You quota upgrade.\",\"request_id\":\"ueut\"}"}】
2022-01-23 23:05:20 | FUNC:circle | UNIT: CallThirdParty[ERR_QUOTA] | TIP:WARNING |TEXT: 【postbackService resp. Error [tip:QUOTA][host:https://api_host.io/api/json/ip/][sender:http://sky.net.kg:1333/rpc][ip:178.176.77.64]】
2022-01-23 23:05:20 | FUNC:GenFilter.ErrDaemon | UNIT: resp.error.counter | TIP:WARNING |TEXT: 【зарегистрировано ошибок:1|предел:2】
2022-01-23 23:05:20 | FUNC:ⓇⓅⒸ | UNIT: http.Post[http://sky.net.kg:1333/rpc] | TIP:RESPONSE |TEXT: 【[body:{"RespStatus":"200 OK","RespCode":200,"RespBody":"{\"success\":false,\"message\":\"You quota upgrade.\",\"request_id\":\"upMMI\"}"}]】
2022-01-23 23:05:20 | FUNC:circle | UNIT: CallThirdParty[ERR_QUOTA] | TIP:WARNING |TEXT: 【postbackService resp. Error [tip:QUOTA][host:https://api_host.io/api/json/ip/][sender:http://sky.net.kg:1333/rpc][ip:5.18.146.234]】
# Запросы больше не отправляются, а следующий хост начнет работу через 18 минут
2022-01-23 23:05:20 | FUNC:GenFilter.replaceSender | UNIT: bad:http://sky.net.kg:1333/rpc|next:http://sky.net.kg:1332/rpc | TIP:INFO |TEXT: 【sleep:18 minutes START】
2022-01-23 23:05:20 | FUNC:circle | UNIT: CallThirdParty[ERR_QUOTA] | TIP:WARNING |TEXT: 【запрос не был отправлен [tip:MOCK][resp: {0 0 2022-01-23 23:05:20.2689993 +0300 MSK m=+14160.765454301  Mozilla/5.0 (Linux; Android 10; HRY-LX1T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.98 Mobile Safari/537.36】
2022-01-23 23:05:20 | FUNC:GenFilter.ErrDaemon | UNIT: http://sky.net.kg:1333/rpc | TIP:INFO |TEXT: 【IsMock:true (игнорируем ошибку)】
```

