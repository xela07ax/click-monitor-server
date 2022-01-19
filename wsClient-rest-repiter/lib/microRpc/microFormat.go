package microRpc

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/client-rest-repiter/model"
)

type Rpc struct {
	Alias    string
	Services map[string]func([]byte) []byte
	Loger    chan<- [4]string
}

func NewRpc(loger chan<- [4]string, services map[string]func([]byte) []byte) *Rpc {
	return &Rpc{Alias: "ⓇⓅⒸ", Loger: loger, Services: services}
}

func (rpc *Rpc) InputMsg(rpcMsg []byte) (resp []byte) {
	rpc.Loger <- [4]string{rpc.Alias, "⚡𝓻𝓮𝓺𝓾𝓮𝓼𝓽⚡", fmt.Sprintf("🅱🅾🅳🆈【%s】", rpcMsg), "HTTP_READ"}
	rawResp := &model.RpcFields{}
	// открывакм сообщение
	microRout := &model.MicroFull{}

	err := json.Unmarshal(rpcMsg, microRout)
	if err != nil {
		err = fmt.Errorf("ошибка чтения RPC: %s |body【%s】", err, rpcMsg)
		rpc.Loger <- [4]string{rpc.Alias, "nil", err.Error(), "1"}
		rawResp.RespStatus = err.Error()
		resp, _ = json.Marshal(rawResp)
		return
	}
	svc, ok := rpc.Services[microRout.Service]
	if ok {
		resp = svc([]byte(microRout.Request))
	} else {
		rawResp.RespStatus = fmt.Sprintf("сервис [%s] ненайден", microRout.Service)
		rpc.Loger <- [4]string{rpc.Alias, "nil", rawResp.RespStatus, "ERROR"}
		return
	}

	return
}
