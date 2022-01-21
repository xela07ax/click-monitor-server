package domain

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"net"
	"net/http"
	"time"
)

type Api struct {
	FuncName string
	Text     string
	Status   int
	Show     bool
	UpdNum   int
}

type DaemonRouter struct {
	listener net.Listener
	cfg      *model.Config
	Loger    chan<- [4]string
}

func NewRouter(config *model.Config, listener net.Listener, loger chan<- [4]string) *DaemonRouter {
	return &DaemonRouter{
		listener: listener,
		Loger:    loger,
		cfg:      config,
	}
}
func (d *DaemonRouter) Start(serveWs, sentWs, homeWs func(w http.ResponseWriter, r *http.Request)) {
	router := mux.NewRouter()
	// Шэринг ресурсов
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	allowedCredentials := handlers.AllowCredentials() //для куков

	server := &http.Server{
		Handler:        handlers.CORS(headersOk, originsOk, methodsOk, allowedCredentials)(router),
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}
	router.HandleFunc("/", homeWs)
	router.HandleFunc("/sentws", sentWs)
	router.HandleFunc("/ws", serveWs)

	go func() {
		err := server.Serve(d.listener)
		if err != nil {
			d.Loger <- [4]string{"Daemon", "Start", fmt.Sprintf("err[server.Serve]:%v", err), "ERROR"}
		}
	}()
	d.Loger <- [4]string{"Front Http Server", "nil", fmt.Sprintf("Server is started")}
}




