package inputRpc

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/rest-repiter/model"
	"io/ioutil"
	"net/http"
	"time"
)

type Rpc struct {
	Alias string
	cfg *model.Config
	Services  map[string]chan <-model.Pino
	Loger chan <-[4]string
}

func NewRpc(cfg *model.Config, loger chan <-[4]string, services map[string]chan <-model.Pino) *Rpc  {
	return &Rpc{Alias: "ⓇⓅⒸ", cfg: cfg, Loger: loger, Services: services}
}

func (rpc *Rpc) InputRpc(w http.ResponseWriter, r *http.Request) {
	rpcMsg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ertx := fmt.Sprintf("COM:Ошибка чтения тела: %s | ERTX:can't read body", err)
		rpc.Loger <- [4]string{rpc.Alias, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusConflict) // 409
		return
	}
	err, resp := rpc.InputMsg(rpcMsg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		ertx := fmt.Sprintf("COM: не удалось отправить ответ | ERTX:%v", err)
		rpc.Loger <- [4]string{rpc.Alias, "nil", ertx, "1"}
		http.Error(w, ertx, http.StatusBadRequest) // 400
		return
	}
	rpc.Loger <- [4]string{rpc.Alias, "⚡𝓼𝓽𝓪𝓽𝓾𝓼 𝟮𝟬𝟬⚡", fmt.Sprintf("🅱🅾🅳🆈【%s】", string(resp)), "HTTP_WRITE"}
}
func (rpc *Rpc) InputMsg(rpcMsg []byte) (err error, resp []byte) {
	rpc.Loger <- [4]string{rpc.Alias, "⚡𝓻𝓮𝓺𝓾𝓮𝓼𝓽⚡", fmt.Sprintf("🅱🅾🅳🆈【%s】", rpcMsg), "HTTP_READ"}
	microRout := &model.Micro{}
	err = json.Unmarshal(rpcMsg, microRout)
	if err != nil {
		err = fmt.Errorf("COM:Ошибка чтения RPC: %s | ERTX:can't read RPC", err)
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		return
	}
	svc, ok := rpc.Services[microRout.Service]
	if !ok {
		err = fmt.Errorf("COM:Ошибка поиска| service: %v | ERTX:can't find Service| use:%v", microRout.Service, rpc.Services)
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		return
	}
	back := make(chan model.Pino)
	select {
	case svc <- model.Pino{
		Message: "",
		Traffic: rpcMsg,
		Door:    back,
	}:
	case <-time.After(30 * time.Second):
		err = fmt.Errorf("COM: Сервис[%s] не принял данные |Err:_timeout_%d_", microRout.Service, 30)
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		return
	}
	select {
	case result := <-back:
		if result.Err != nil {
			err = fmt.Errorf("COM: Сервис[%s] ответил с ошибкой |Err:%v", microRout.Service, result.Err)
			rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
			return
		}
		resp = result.Traffic
	case <-time.After(30 * time.Second):
		err = fmt.Errorf("COM: Сервис[%s] не ответил |Err:_timeout_%d_", microRout.Service, 30)
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		return
	}
	return
}