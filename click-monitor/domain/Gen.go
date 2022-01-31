package domain

import (
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/clickhouse"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/db"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/reporter"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/reqestWrkHub"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/timeBetweenGen"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"github.com/xela07ax/toolsXela/tp"
	"log"
	"time"
)

type GenFilter struct {
	Interval               time.Duration
	currentTimestamp       time.Time
	minusIntervalTimestamp time.Time
	ticker                 <-chan time.Time
	filter                 model.LogFilter
	chRepo                 clickhouse.LogRepo
	cfg                    *model.Config
	reporting              *reporter.Reporting
	db                     *db.Slowpoke
	ErrReq                 chan interface{}
	sended                 map[string]interface{} // будем кешировать отправленных в работу (осторошно, возможна утечка памяти!)
	globalErr              int
	//iSender                error
	Sender      *model.Sender
	limit       int
	cashSenders map[*model.Sender]interface{}
	Loger       chan<- [4]string
}

func New_GenFilter_ChDbMonitor(cfg *model.Config, loger chan<- [4]string) *GenFilter {
	slowPoke := db.NewStore(cfg, loger)
	slowPoke.RunMinion()

	gen := &GenFilter{db: slowPoke, reporting: reporter.NewReporting(cfg, loger),
		Interval: time.Duration(cfg.Interval) * time.Second,
		Loger:    loger, chRepo: clickhouse.NewLogRepo(cfg, loger),
		cfg: cfg, ErrReq: make(chan interface{}, 100),
		cashSenders: make(map[*model.Sender]interface{}),
		sended:      make(map[string]interface{})}

	if cfg.ModeStart.Counter {
		loger <- [4]string{"StartMonitor", "startMode.Counter[TableIpAddress]", fmt.Sprintf("%d", len(slowPoke.TableIpAddress.ReadAll())), "INFO"}
		tp.ExitWithSecTimeout(0)
	}
	if cfg.ModeStart.Updater {
		// gen.ReaDatabase()
		gen.UpdateDatabase()
		loger <- [4]string{"StartMonitor", "startMode.Updater[TableIpAddress]", "IS OK", "INFO"}
		tp.ExitWithSecTimeout(0)
	}

	if cfg.ModeStart.ClickMonitorAsync {
		// функция которая определин соттветствует ли ответ ожидаемому
		// хаб для воркеров
		sendersHub := reqestWrkHub.NewHub(cfg, Validater, loger)
		toRequest := sendersHub.Input // канал для отправки воркеру (который возьмется)
		// каналы отчетов, на основе err реквеста и validater() == true
		respOkCh := sendersHub.LogIpqs
		respErCh := sendersHub.LogBad

		// инициализируем фоновый таймер, который сделает нам еще и даты для фильтра в sql
		bTimer := timeBetweenGen.NewBetweenTimer(cfg.ChDb.Timezone, cfg.Interval, loger)
		// канал для приема уведомления о начале нового sql запроса с заданной between
		tickSql := bTimer.Ticket
		// инициализируем воркера который будет собирать сам запрос и обрабатывать выдачу с базы
		go gen.DaemonSqlMaker(tickSql, toRequest)
		go gen.RequestDaemonWriter(respOkCh, respErCh)
		bTimer.StartDaemonTicker()
		return gen
	}
	log.Fatal("не выбран ни один режим работы")
	return nil
}
