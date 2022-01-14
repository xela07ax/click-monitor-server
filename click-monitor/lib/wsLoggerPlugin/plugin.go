package wsLoggerPlugin

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func Logx(txt string)  {
	fmt.Printf("[LOGIX]:%s\n", txt)
}
func NewWsLogger() *WsLogger {
	return &WsLogger{
		broadcast:  make(chan []byte),
		Input:      make(chan []byte, 100),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		Loger:      nil,
	}
}

// serveWs handles websocket requests from the peer.
func (hub *WsLogger) ServeWs(w http.ResponseWriter, r *http.Request) {
	hub.Loger <- [4]string{"WsLogger", "ServeWs", "input http client"}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.Loger <- [4]string{"WsLogger", "ServeWs", fmt.Sprintf("[upgrader.Upgrade]err:%v", err), "ERROR"}
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client
	// Разрешить сбор памяти, на которую ссылается вызывающий абонент, выполнив всю работу в
	// новых goroutines.
	go client.writePump()
	go client.readPump()
}

// serveWs handles websocket requests from the peer.
func (hub *WsLogger) SentWS(w http.ResponseWriter, r *http.Request) {
	hub.Loger <- [4]string{"WsLogger", "SentWS", "input http client"}
	msgRaw := readBodySimple(w, r)
	if len(msgRaw) == 0 {
		return
	}
	// Реализация Writer-а
	log.Printf("--W-true> отправили:%s\n", msgRaw)
	ertx := "-sendMsg->true"
	resp(w, r, "sendMsg", ertx, 0, true)
}

func (hub *WsLogger) HomePageWs(w http.ResponseWriter, r *http.Request) {
	err := homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
	if err != nil {
		if hub.Loger != nil {
			hub.Loger <- [4]string{"Daemon", "HomePageWs", fmt.Sprintf("err[homeTemplate.Execute]:%v", err), "ERROR"}
		}
	}
}

var homeTemplate = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<title>Chat Example</title>
<script type="text/javascript">
window.onload = function () {
    var conn;
    var msg = document.getElementById("msg");
    var log = document.getElementById("log");

    function appendLog(item) {
        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }

    document.getElementById("form").onsubmit = function () {
        if (!conn) {
            return false;
        }
        if (!msg.value) {
            return false;
        }
        conn.send(msg.value);
        msg.value = "";
        return false;
    };

    if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + "/ws");
        conn.onclose = function (evt) {
            var item = document.createElement("div");
            item.innerHTML = "<b>Connection closed.</b>";
            appendLog(item);
        };
        conn.onmessage = function (evt) {
            var messages = evt.data.split('\n');
            for (var i = 0; i < messages.length; i++) {
                var item = document.createElement("div");
                item.innerText = messages[i];
                appendLog(item);
            }
        };
    } else {
        var item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
    }
};
</script>
<style type="text/css">
html {
    overflow: hidden;
}

body {
    overflow: hidden;
    padding: 0;
    margin: 0;
    width: 100%;
    height: 100%;
    background: gray;
}

#log {
    background: white;
    margin: 0;
    padding: 0.5em 0.5em 0.5em 0.5em;
    position: absolute;
    top: 0.5em;
    left: 0.5em;
    right: 0.5em;
    bottom: 3em;
    overflow: auto;
}

#form {
    padding: 0 0.5em 0 0.5em;
    margin: 0;
    position: absolute;
    bottom: 1em;
    left: 0px;
    width: 100%;
    overflow: hidden;
}

</style>
</head>
<body>
<div id="log"></div>
<form id="form">
    <input type="submit" value="Send" />
    <input type="text" id="msg" size="64" autofocus />
</form>
</body>
</html>
`))
