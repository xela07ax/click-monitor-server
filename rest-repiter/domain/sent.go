package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xela07ax/rest-repiter/model"
	"github.com/xela07ax/toolsXela/tp"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	redirectUrl = "http://localhost:13338"
	historyFile = "./lastRedirect.json"
)

var History []model.Handle

func Sent(req model.Handle, out chan<- model.Pino, loger chan [4]string) {
	loger <- [4]string{"Sent", req.Method, fmt.Sprintf("%v", req)}
	var resp *http.Response
	var err error

	url := req.RedirectUrl
	if req.Params != "" {
		url = fmt.Sprintf("%s?%s", url, req.Params)
	}
	fmt.Println("req.Send")
	fmt.Println(req.Method)
	if req.Send {
		switch req.Method {
		case http.MethodPost:
			r := bytes.NewReader([]byte(req.Body))
			resp, err = http.Post(url, req.Type, r)
			if err != nil {
				loger <- [4]string{"Sent", req.Method, fmt.Sprintf("err:%v|url:%s|body:%s", err, url, req.Body), "ERROR"}
			}
		case http.MethodGet:
			fmt.Println("http.Post")
			resp, err = http.Get(url)
			fmt.Println(resp)
			if err != nil {
				loger <- [4]string{"Sent", req.Method, fmt.Sprintf("err:%v|url:%s|body:%s", err, url, req.Body), "ERROR"}
			}
		}
	} else {
		loger <- [4]string{"Sent", req.Method, fmt.Sprintf("url:%s|body:%s", url, req.Body), "MOCK"}

	}

	var toDoor model.RpcFields
	if resp != nil {
		defer resp.Body.Close()
		// io.Copy(os.Stdout, resp.Body)
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			loger <- [4]string{"Sent", "get[ioutil.ReadAll]", fmt.Sprintf("err:%v|status:%s|code:%d|url:%s|body:%s", err, resp.Status, resp.StatusCode, url, req.Body), "ERROR"}
		}
		toDoor = model.RpcFields{
			RespStatus: resp.Status,
			RespCode:   resp.StatusCode,
			RespBody:   string(body),
		}
		loger <- [4]string{"Sent", url, fmt.Sprintf("%v", toDoor), "RESPONSE"}
	}
	dat, err := json.Marshal(toDoor)
	if err != nil {
		loger <- [4]string{"Sent", "get[json.Marshal]", fmt.Sprintf("err:%v|status:%s|code:%d|url:%s|toDoor:%v", err, resp.Status, resp.StatusCode, url, toDoor), "ERROR"}
	}
	out <- model.Pino{Traffic: dat}
}
const panicMsg = "детектирована паника"

func recovery(uagDoor chan model.Pino, loger chan [4]string) {
	if rMsg := recover(); rMsg != nil { // Если была паника, будем отвечать ошибкой в сентри и лог
		err := fmt.Errorf("worker[RpcSentWorker]info[%s]err:%v", panicMsg, rMsg)
		loger <- [4]string{"Handler", "json.Unmarshal", fmt.Sprintf("err:%v", err), "PANIC"}
		time.Sleep(5 * time.Second)
		RunRpcSentWorker(uagDoor, loger)
	}
}
func RunRpcSentWorker(uagDoor chan model.Pino, loger chan [4]string) {
	defer final(loger)
	defer recovery(uagDoor, loger)
	for {
		pino := <-uagDoor
		body := new(model.Body)
		err := json.Unmarshal(pino.Traffic, body)
		if err != nil {
			loger <- [4]string{"Handler", "json.Unmarshal", fmt.Sprintf("err:%v|data:%s", err, pino.Traffic), "ERROR"}
		}
		handle := new(model.Handle)
		err = json.Unmarshal([]byte(body.Request), handle)
		if err != nil {
			loger <- [4]string{"Handler", "json.Unmarshal", fmt.Sprintf("err:%v|data:%s", err, pino.Traffic), "ERROR"}
		}
		fmt.Printf("[HANDLER]handle:%v\n", handle)
		if err != nil {
			loger <- [4]string{"Handler", "json.Unmarshal", fmt.Sprintf("err:%v|data:%s", err, pino.Traffic), "ERROR"}
		}
		History = append(History, *handle)
		Sent(*handle, pino.Door, loger)
	}
}
func final(loger chan [4]string) {
	subNmae := "Domain=>Sent/Handler=Closer"
	loger <- [4]string{subNmae, "final", "Безопасное завершение программы"}
	if len(History) < 1 {
		loger <- [4]string{subNmae, "len(History)", "Нет данных для сохранения"}
		return
	}
	f, err := tp.CreateOpenFile(historyFile)
	if err != nil {
		loger <- [4]string{subNmae, "tp.CreateOpenFile", err.Error(), "ERROR"}
		return
	}
	defer f.Close()
	loger <- [4]string{subNmae, "f.Write", fmt.Sprintf("сохраняем [%d] запросов", len(History))}

	data, err := json.Marshal(History)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(data)
	if err != nil {
		panic(err)
	}
	loger <- [4]string{subNmae, "END", "Шикарно"}
}
