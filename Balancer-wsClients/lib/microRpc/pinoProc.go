package microRpc

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/rest-repiter/model"
	"io/ioutil"
	"net/http"
)

type Rpc struct {
	services map[string]map[string]func([]byte)[]byte
	Alias string
	cfg *model.Config
	Loger chan <-[4]string
}

func NewRpc(cfg *model.Config, loger chan <-[4]string, services map[string]map[string]func([]byte)[]byte) *Rpc  {
	return &Rpc{Alias: "ⓇⓅⒸ", cfg: cfg, Loger: loger, services: services}
}


func (rpc *Rpc) InputRpcServe(w http.ResponseWriter, r *http.Request) {
	rpcMsg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ertx := fmt.Sprintf("comm:Ошибка чтения тела: %s | ERTX:can't read body", err)
		rpc.Loger <- [4]string{rpc.Alias, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusConflict) // 409
		return
	}
	resp := rpc.InputMsg(rpcMsg)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		ertx := fmt.Sprintf("comm: не удалось отправить ответ | ERTX:%v", err)
		rpc.Loger <- [4]string{rpc.Alias, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusBadRequest) // 400
		return
	} else {
		rpc.Loger <- [4]string{rpc.Alias, "⚡𝓼𝓽𝓪𝓽𝓾𝓼 𝟮𝟬𝟬⚡", fmt.Sprintf("🅱🅾🅳🆈【%s】", string(resp)), "HTTP_WRITE"}
	}
}

func (rpc *Rpc) InputMsg(rpcMsg []byte) (resp []byte) {
	rpc.Loger <- [4]string{rpc.Alias, "⚡𝓻𝓮𝓺𝓾𝓮𝓼𝓽⚡", fmt.Sprintf("🅱🅾🅳🆈【%s】", rpcMsg), "HTTP_READ"}
	microRout := &model.Micro{}
	err := json.Unmarshal(rpcMsg, microRout)
	if err != nil {
		example, _ := json.Marshal(model.Micro{
			Service:  "your.service",
			Endpoint: "service.endpoint",
		})
		err = fmt.Errorf("comm:Ошибка чтения RPC: %s |msg【%s】example 【%s】", err,rpcMsg, example)
		resp = []byte(err.Error())
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		return
	}
	_, ok := rpc.services[microRout.Service]
	if !ok {
		err = fmt.Errorf("comm:Ошибка поиска сервиса: %v | use:%v", microRout.Service, rpc.services)
		resp = []byte(err.Error())
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		return
	}
	endpoint, ok := rpc.services[microRout.Service][microRout.Endpoint]
	if !ok {
		err = fmt.Errorf("comm:Ошибка поиска целевого метода: %v | use:%v", microRout.Service, rpc.services[microRout.Service])
		resp = []byte(err.Error())
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		return
	}
	microRequest := &model.MicroRequest{}
	err = json.Unmarshal(rpcMsg, microRequest)
	if err != nil {
		err = fmt.Errorf("comm:Ошибка чтения поля запроса RPC: %s", err)
		resp = []byte(err.Error())
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		return
	}

	resp = endpoint([]byte(microRequest.Request))
	if len(resp) == 0 {
		err = fmt.Errorf("comm:Ошибка ответа сервиса RPC: ответ не может быть пустым")
		resp = []byte(err.Error())
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		return
	}
	return resp
}