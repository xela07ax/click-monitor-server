package repiter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xela07ax/client-rest-repiter/model"
	"io/ioutil"
	"net/http"
)


var History []model.Handle

type Repiter struct {
	Loger chan [4]string
}

func NewRepiter(loger chan [4]string) *Repiter {
	return &Repiter{loger}
}

func (r *Repiter) RecieveHandleMsg(handleMsg []byte) []byte {
	handle := new(model.Handle)
	err := json.Unmarshal(handleMsg, handle)
	if err != nil {
		r.Loger <- [4]string{"Repiter", "RecieveHandleMsg[json.Unmarshal]", fmt.Sprintf("err:%v|data:%s", err, handleMsg), "ERROR"}
	}
	History = append(History, *handle)
	return exec(*handle, r.Loger)
}

func exec(req model.Handle, loger chan [4]string) []byte {
	loger <- [4]string{"Sent", req.Method, fmt.Sprintf("%v", req)}
	var resp *http.Response
	var err error

	url := req.RedirectUrl
	if req.Params != "" {
		url = fmt.Sprintf("%s?%s", url, req.Params)
	}
	if req.Send {
		switch req.Method {
		case http.MethodPost:
			r := bytes.NewReader([]byte(req.Body))
			resp, err = http.Post(url, req.Type, r)
			if err != nil {
				loger <- [4]string{"Sent", req.Method, fmt.Sprintf("err:%v|url:%s|body:%s", err, url, req.Body), "ERROR"}
			}
		case http.MethodGet:
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
	return dat
}
