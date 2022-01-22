package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"io/ioutil"
	"net/http"
	"time"
)

const subName = "ⓇⓅⒸ"

func RpcRequest(service, data, url string, loger chan<- [4]string) (response *model.RpcFields, err error)  {
	microRout := &model.RpcModel{
		Micro:      model.Micro{ Service: service },
		Body:       model.Body{ Request: data },
	}
	body, err := json.Marshal(microRout)
	if err != nil {
		loger <- [4]string{ subName, "json.Marshal", fmt.Sprintf("err:%v|url:%s|microRout:%v", err, url, microRout), "ERROR"}
		return nil, err
	}
	r := bytes.NewReader(body)
	//loger <- [4]string{ subName, "http.Post", fmt.Sprintf("url:%s|body:%s", url, body), "REQUEST"}
	resp, err := http.Post(url, "", r)
	if err != nil {
		loger <- [4]string{ subName, fmt.Sprintf("http.Post[%s]", url), fmt.Sprintf("err:%v|url:%s|body:%s", err, url, data), "ERROR"}
		return nil, err
	}
	defer resp.Body.Close()

	var repBody []byte
	repBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		loger <- [4]string{ subName, "ioutil.ReadAll", fmt.Sprintf("resp.Body error: %v", err), "ERROR"}
		return nil, err
	}

	if len(repBody) > 0 {
		loger <- [4]string{ subName, "http.Post", fmt.Sprintf("body:%s", repBody), "RESPONSE"}
		response = new(model.RpcFields)
		err = json.Unmarshal(repBody, response)
		if err != nil {
			loger <- [4]string{ subName, "json.Unmarshal", fmt.Sprintf("resp.Body error: %v", err), "ERROR"}
			return nil, err
		}

		fmt.Printf("http.Post_ok:%v\n", response)
	}

	return
}

type Rpc struct {
	Services           map[string]chan<- model.Pino
	timeoutWaitService time.Duration
	Loger              chan<- [4]string
}

func NewRpcServiceHandler(timeoutWaitService int, loger chan<- [4]string) *Rpc {
	return &Rpc{
		Services:           make(map[string]chan<- model.Pino),
		timeoutWaitService: time.Duration(timeoutWaitService) * time.Second,
		Loger:              loger,
	}
}
func (app *Rpc) PoolClientCreate(w http.ResponseWriter, r *http.Request) {
	//app.LogerChan <- [4]string{subName, "new_request", "new http input request", "HTTP_HELLO"}
	//		service = 'go.tracker.svc.capproc', <- наш маршрут
	//		endpoint = 'Capproc.HandleConversion',

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ertx := fmt.Sprintf("COM:Ошибка чтения тела: %s | ERTX:can't read body", err)
		app.Loger <- [4]string{subName, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusConflict) // 409
		return
	}

	app.Loger <- [4]string{subName, "⚡𝓻𝓮𝓺𝓾𝓮𝓼𝓽⚡", fmt.Sprintf("🅱🅾🅳🆈【%s】", b), "HTTP_READ"}
	microRout := &model.Micro{}
	err = json.Unmarshal(b, microRout)
	if err != nil {
		ertx := fmt.Sprintf("COM:Ошибка чтения RPC: %s | ERTX:can't read RPC", err)
		app.Loger <- [4]string{subName, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusConflict)
		return
	}
	svc, ok := app.Services[microRout.Service]
	if !ok {
		ertx := fmt.Sprintf("COM:Ошибка поиска| service: %v | ERTX:can't find Service| use:%v", microRout.Service, app.Services)
		app.Loger <- [4]string{subName, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusNotFound)
		return
	}
	back := make(chan model.Pino)
	select {
	case svc <- model.Pino{
		Message: "",
		Traffic: b,
		Door:    back,
	}:
	case <-time.After(app.timeoutWaitService):
		ertx := fmt.Sprintf("COM: Сервис[%s] не принял данные |Err:_timeout_%d_", microRout.Service, app.timeoutWaitService)
		app.Loger <- [4]string{subName, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusBadRequest)
		return
	}
	var dat []byte
	select {
	case resp := <-back:
		if resp.Err != nil {
			ertx := fmt.Sprintf("COM: Сервис[%s] ответил с ошибкой |Err:%v", microRout.Service, resp.Err)
			app.Loger <- [4]string{subName, "nil", ertx, "1"}
			http.Error(w, ertx, http.StatusBadRequest)
			return
		}
		dat = resp.Traffic
	case <-time.After(app.timeoutWaitService * time.Second):
		ertx := fmt.Sprintf("COM: Сервис[%s] не ответил |Err:_timeout_%d_", microRout.Service, app.timeoutWaitService)
		app.Loger <- [4]string{subName, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusBadRequest)
		return
	}
	// "resp.body": "{\"id\":\"error.server_internal\",\"code\":500,\"status\":\"Internal Server Error\"}",
	// "resp.status": 500
	// or
	// "resp.body": "{\"url\":\"https:\/\/www.rarlab.com\",\"client_id\":\"a9525895-c726-5acc-a21b-f1b365dc8c8e\",\"session_id\":\"2700ffee-003c-5b13-b940-16e7fccc4692\",\"click_id\":\"c6196510-875b-590a-8cce-671d59bfaa88\"}",
	// "resp.status": 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(dat)
	if err != nil {
		ertx := fmt.Sprintf("COM: Сервис[%s] не удалось отправить ответ | ERTX:%v", microRout.Service, err)
		app.Loger <- [4]string{subName, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusBadRequest) // 400
		return
	}
	app.Loger <- [4]string{subName, "⚡𝓼𝓽𝓪𝓽𝓾𝓼 𝟮𝟬𝟬⚡", fmt.Sprintf("🅱🅾🅳🆈【%s】", dat), "HTTP_WRITE"}
}
