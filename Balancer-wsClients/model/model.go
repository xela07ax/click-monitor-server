package model


type MicroRequest struct {
	Request string `json:"request"`
}

type Micro struct {
	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
}

type RpcModel struct {
	Micro
	MicroRequest
}

type MicroExec struct {
	Param  string `json:"param"`
	Request string `json:"request"`
}

type Pino struct {
	Message string
	Err     error
	Traffic []byte
	Door    chan Pino
}


