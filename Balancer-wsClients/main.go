package main

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/rest-repiter/domain"
	"github.com/xela07ax/rest-repiter/lib/chLogger"
	"github.com/xela07ax/rest-repiter/lib/hubWsServices"
	"github.com/xela07ax/rest-repiter/lib/microRpc"
	"github.com/xela07ax/rest-repiter/lib/wsLoggerPlugin"
	"github.com/xela07ax/rest-repiter/model"
	"github.com/xela07ax/toolsXela/tp"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const configName = "config.json"

func main() {
	// Подготовим конфиг
	dir, err := tp.BinDir()
	tp.Fck(err)
	cfgPth := filepath.Join(dir, configName)
	fmt.Printf("config path: %s\n", cfgPth)
	cfg := newConfig(cfgPth)
	cfg.Path.BinDir = dir
	cfg.ChExit = make(chan os.Signal, 2)

	routes := make(map[string]func(w http.ResponseWriter, r *http.Request))
	services := make(map[string]map[string]func([]byte)[]byte)

	// Создаем логер
	logEr := chLogger.NewChLoger(&chLogger.Config{Dir: cfg.Path.Log})
	logEr.RunMinion()

	// {"endpoint": "GetListClients","service": "go.tracker.svc.hubServices"}
	myrpc := microRpc.NewRpc(cfg, logEr.ChInLog, services)
	routes["/rpc"] = myrpc.InputRpcServe

	// Добавляем ws терминал куда будут сыпаться логи и отправлять команды API
	wsLogger := wsLoggerPlugin.NewWsLogger(myrpc.InputMsg)
	go wsLogger.Run()
	logEr.Options.Broadcast = wsLogger.Input
	routes["/"] = wsLogger.HomePageWs
	//routes["/sentws"] = wsLogger.SentWS
	routes["/ws"] = wsLogger.ServeWs

	// Добавляем Hub для клиентов воркеров
	connectorWs := hubWsServices.NewWsConnector(myrpc.InputMsg, logEr.ChInLog)
	go connectorWs.Run()
	routes["/client/ws"] = connectorWs.ServeWs

	services["go.tracker.svc.hubServices"] = make(map[string]func([]byte)[]byte)
	services["go.tracker.svc.hubServices"]["GetListClients"] = connectorWs.GetListClients
	services["go.tracker.svc.hubServices"]["SentClientMsg"] = connectorWs.SentMsgFromClient
	services["go.tracker.svc.repiter"] = make(map[string]func([]byte)[]byte)
	services["go.tracker.svc.repiter"][""] = connectorWs.SentMsgFromAnyClient
	// localhost 【::1】

	//services := make(map[string]chan<- model.Pino)
	//rpc := inputRpc.NewRpc(cfg, logErWs.Loger, services)
	//logErWs.Interpretator = rpc.InputMsg

	//slowPoke := db.NewStore(cfg, loger)
	//slowPoke.RunMinion()
	// инициализируем репозитарии
	//reporting := reporter.NewReporting(cfg, logEr.ChInLog)

	// Запускаем сервис
	err = domain.RunDaemon(cfg, routes,  logEr.ChInLog, nil)
	if err != nil {
		logEr.ChInLog <- [4]string{"Main", "domain.RunDaemon", fmt.Sprintf("%v", err), "ERROR"}
	}
	fmt.Println("▌ █║ exiting after 2 second │║")
	time.Sleep(1 * time.Second)
}

func newConfig(configDir string) *model.Config {
	// Открываем конфиг
	fi, err := tp.OpenReadFile(configDir)
	if err != nil {
		fmt.Printf("Ошибка при открытии конфигурации %s: %s\n", configDir, err)
		tp.ExitWithSecTimeout(1)
	}
	config := new(model.Config)
	err = json.Unmarshal(fi, config)
	if err != nil {
		fmt.Printf("Ошибка чтения JSON %s: %s\n", configDir, err)
		tp.ExitWithSecTimeout(1)
	}
	return config
}
