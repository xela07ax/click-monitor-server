package main

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/rest-repiter/domain"
	"github.com/xela07ax/rest-repiter/lib/chLogger"
	"github.com/xela07ax/rest-repiter/lib/inputRpc"
	"github.com/xela07ax/rest-repiter/lib/wsLoggerPlugin"
	"github.com/xela07ax/rest-repiter/model"
	"github.com/xela07ax/toolsXela/tp"
	"os"
	"path/filepath"
	"time"
)

const configName = "config.json"

func main() {
	fmt.Printf("ಠ┗(▀̿Ĺ̯▀̿ ̿)┓ \n")
	fmt.Printf("        Click monitor v1.12 +repiter module v1.4\n")
	fmt.Printf("      ٩◔̯◔۶\n")
	// Подготовим конфиг
	dir, err := tp.BinDir()
	tp.Fck(err)
	cfgPth := filepath.Join(dir, configName)
	fmt.Printf("config path: %s\n", cfgPth)
	cfg := newConfig(cfgPth)
	cfg.Path.BinDir = dir
	cfg.ChExit = make(chan os.Signal,2)

	logErWs := wsLoggerPlugin.NewWsLogger()
	go logErWs.Run()
	// Создаем логер
	logEr := chLogger.NewChLoger(&chLogger.Config{Dir: cfg.Path.Log, Broadcast: logErWs.Input})
	logEr.RunMinion()
	logErWs.Loger = logEr.ChInLog

	services := make(map[string]chan <-model.Pino)
	rpc := inputRpc.NewRpc(cfg, logErWs.Loger, services)
	logErWs.Interpretator = rpc.InputMsg

	// Инициализируем репозитории
	domain.New_GenFilter_ChDbMonitor(cfg, logEr.ChInLog)

	// Запускаем сервис
	err = domain.RunDaemon(cfg, logEr.ChInLog, logErWs.ServeWs, logErWs.SentWS, logErWs.HomePageWs, nil)
	if err != nil {
		logEr.ChInLog <- [4]string{"Main", "domain.RunDaemon", fmt.Sprintf("%v", err), "ERROR"}
	}
	fmt.Println("▌ █║ exiting after 2 second │║")
	time.Sleep(1*time.Second)
}


func newConfig(configDir string) *model.Config {
	// Открываем конфиг
	fi,err := tp.OpenReadFile(configDir)
	if err != nil {
		fmt.Printf("Ошибка при открытии конфигурации %s: %s\n",configDir,err)
		tp.ExitWithSecTimeout(1)
	}
	config := new(model.Config)
	err = json.Unmarshal(fi, config)
	if err != nil {
		fmt.Printf("Ошибка чтения JSON %s: %s\n",configDir,err)
		tp.ExitWithSecTimeout(1)
	}
	return config
}

