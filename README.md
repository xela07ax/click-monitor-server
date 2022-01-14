## Click Monitor and Repiter Tools 
#### Отправка запросов с удаленного сервера


1) Считывает сообщения из БД о новых кликах
    - мониторинг таблицы на появление новых кликов
2) Подготавливает URL для отправки в сторонний сервис
    - подготавливает, но сам не отправляет, так как его IP не предназначен для таких запросов
3) Отправляет на последующую отправку в ***rest-repiter*** который на стороннем сервере ждет rest сообщения 
4) Сортирует полученный результат на ошибки, ок и найденные в кеше для мини Репортинга 
    - отчет в реальном времени в папке report
    - сохранение в OLTP базу для дальнейшей работы с данными
5) Кеширует запросы по ключу IP + UserAgent

Скриншот ***click-monitor.exe***
![Скриншот click-monitor.exe](./docs/desctop_app.png)  
<img src="./docs/report_system.png" width="550" />  
Скриншот удаленного ws терминала ***rest-repiter.exe***  
<img src="./docs/ws_logger-repiter.png" width="550" />  

Компиляция:  
```sh
CGO_ENABLED=0 GOOS=windows go build -gcflags "all=-N -l" -o service.exe
CGO_ENABLED=0 GOOS=linux go build -gcflags "all=-N -l" -o service
```
Конфигурации лежат в папке bin 

Запуск:  
```sh
./rest-repiter_v1.4 1>log1stdout.txt 2>log1stderr.txt
```
