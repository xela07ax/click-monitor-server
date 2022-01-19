package main

import (
	"fmt"
	"github.com/xela07ax/client-rest-repiter/lib/microRpc"
	"github.com/xela07ax/client-rest-repiter/lib/repiter"
	"github.com/xela07ax/client-rest-repiter/lib/wsHandler"
	"github.com/xela07ax/toolsXela/chLogger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	hostServerDefault = "localhost:1331"
	rpcRepiterDefault = "go.tracker.svc.repiter"
)

func main() {
	// Создаем логер
	logEr := chLogger.NewChLoger(&chLogger.Config{})
	logEr.RunLogerDaemon()

	// инициализируем microRPC для роутинга входящих сообщений между демонами
	services := make(map[string]func([]byte) []byte)
	miRpc := microRpc.NewRpc(logEr.ChInLog, services)

	// инициализируем вебсокет соединение с сервером
	hostServer := hostServerDefault
	if len(os.Args) > 1 {
		hostServer = os.Args[1]
		logEr.ChInLog <- [4]string{"main", "args", fmt.Sprintf("параметры хоста сервера переданы: %s", hostServer), "INFO"}
	} else {
		logEr.ChInLog <- [4]string{"main", "args", fmt.Sprintf("параметры не переданы, стандартный хост: %s", hostServer), "WARNING"}
	}
	wsHandler.NewWsDaemon(hostServer, miRpc.InputMsg, logEr.ChInLog)

	// модуль полезного действия. отправляет запросы, которые были переданы в программу
	repit := repiter.NewRepiter(logEr.ChInLog)
	services[rpcRepiterDefault] = repit.RecieveHandleMsg
	logEr.ChInLog <- [4]string{"main", "args", fmt.Sprintf("подключен rpc сервис: %s", rpcRepiterDefault), "INFO"}

	// ожидающие завершения, завершающие функции
	waitForSignal(logEr.ChInLog)
	fmt.Println("(っ◔◡◔)っ ♥ good by.,_,.-")
}

func waitForSignal(loger chan<- [4]string) {
	chExit := make(chan os.Signal, 2)
	signal.Notify(chExit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sig := <-chExit
	loger <- [4]string{"waitForSignal",  fmt.Sprintf("sig[%s]", sig),"безопасное завершение программы | after: 0.5 sec"}
	time.Sleep(500 * time.Millisecond)
}
