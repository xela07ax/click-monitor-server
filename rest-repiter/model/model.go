package model

import "time"

type Body struct {
	Request string `json:"request"`
}

type Micro struct {
	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
}

type RpcFields struct {
	RespStatus string
	RespCode   int
	RespBody   string
}

type Pino struct {
	Message string
	Err     error
	Traffic []byte
	Door    chan Pino
}

type Handle struct {
	Time                                    time.Time
	Send                                    bool
	RedirectUrl, Params, Method, Body, Type string
}
