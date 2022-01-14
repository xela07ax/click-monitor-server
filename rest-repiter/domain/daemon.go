package domain

import (
	"fmt"
	"github.com/xela07ax/rest-repiter/model"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func RunDaemon(config *model.Config, loger chan [4]string, serveWs, sentWs, homeWs func(w http.ResponseWriter, r *http.Request), close chan bool) error {
	// Начинаем открывать сервер
	loger <- [4]string{"Daemon", "nil", fmt.Sprintf("█║ Starting HTTP Listener ▌│║ on port: %d\n", config.HttpServerPort)}
	// Настраиваем сервер
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", config.HttpServerPort))
	if err != nil {
		return fmt.Errorf("daemon[creating listener]err: %v", err)
	}


	sentDoor := make(chan model.Pino)
	rpc := NewRpcServiceHandler(config.Timeout, loger)
	rpc.Services[config.Service] = sentDoor
	go RunRpcSentWorker(sentDoor, loger)

	router := NewRouter(config, l, loger)
	router.Start(rpc.PoolClientCreate, serveWs, sentWs, homeWs)
	waitForSignal(config.ChExit, loger, close)
	return nil
}

func waitForSignal(ChExit chan os.Signal, loger chan<- [4]string, close chan bool) {
	subNmae := "=>Daemon=>Closer"
	signal.Notify(ChExit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	// signal.Notify(ChExit, os.Interrupt)
	loger <- [4]string{"Daemon","waitForSignal","Wait got signal: exiting"}
	s := <-ChExit
	fmt.Printf("%s | [%v] Безопасное завершение программы\n", subNmae, s)
	// Диплога могло вообще не создасться
	// !!! Мы не проверяем записаны ли все данные, просто закрываем это дело
	// loger <- [4]string{"Front Http Server","nil",fmt.Sprintf("Got signal: %v, exiting.", s)}
	// Пока программа при обрыве завершается через этот блок
	loger <- [4]string{"Front Http Server", "nil", fmt.Sprintf("%s | COM:Безопасное завершение программы | DT:%s", subNmae, s)}
	fmt.Printf("Закрываем\n")
	// ReactBro.Close()
	if close != nil {
		close <- true
	}

}
