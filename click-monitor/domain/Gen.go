package domain

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/clickhouse"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/db"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/reporter"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"github.com/xela07ax/toolsXela/tp"
	"strings"
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
	globalErr              int
	//iSender                error
	Sender                 *model.Sender
	limit int
	cashSenders map[*model.Sender]interface{}
	Loger chan<- [4]string
}

func New_GenFilter_ChDbMonitor(cfg *model.Config, loger chan<- [4]string) *GenFilter {
	slowPoke := db.NewStore(cfg, loger)
	slowPoke.RunMinion()

	gen := &GenFilter{db: slowPoke, reporting: reporter.NewReporting(cfg, loger), Interval: time.Duration(cfg.Interval) * time.Second, Loger: loger, chRepo: clickhouse.NewLogRepo(cfg, loger), cfg: cfg, ErrReq: make(chan interface{}, 100), cashSenders: make(map[*model.Sender]interface{})}
	gen.ErrDaemon()
	gen.calc()

	go gen.circle()
	return gen
}

func (g *GenFilter) getIndexSender(sender *model.Sender) int {
	for i, s := range g.cfg.Senders {
		if s == sender {
			return i
		}
	}
	panic(fmt.Errorf("не найдена ссылка отправителя"))
}

func (g *GenFilter) replaceSender() {
	g.cfg.Mock = true
	go func() {
		iSender := g.getIndexSender(g.Sender)
		iSenderNext := iSender + 1
		if iSenderNext == len(g.cfg.Senders) {
			iSenderNext = 0
		}
		nextSender := g.cfg.Senders[iSenderNext]
		g.Loger <- [4]string{"GenFilter.replaceSender", fmt.Sprintf("bad:%s|next:%s", g.Sender.HostRepiter, nextSender.HostRepiter), fmt.Sprintf("sleep:%d minutes START", nextSender.Sleep), "INFO"}
		time.Sleep(time.Duration(nextSender.Sleep) * time.Minute)
		g.Loger <- [4]string{"GenFilter.replaceSender", fmt.Sprintf("bad:%s|next:%s", g.Sender.HostRepiter, nextSender.HostRepiter), fmt.Sprintf("sleep:%d minutes END", nextSender.Sleep), "INFO"}
		if _, ok := g.cashSenders[nextSender]; ok {
			g.limit = nextSender.ThinkRq
			delete(g.cashSenders, nextSender)
		} else {
			g.limit = nextSender.FirstRq
			g.cashSenders[nextSender] = struct {}{}
		}
		g.Sender = nextSender
		g.cfg.Mock = false
	}()
}
func (g *GenFilter) ErrDaemon() {
	if len(g.cfg.Senders) < 1 {
		panic(fmt.Errorf("хостов отправителей не может быть меньше 1-го"))
	}
	g.Sender = g.cfg.Senders[0]
	g.limit = g.Sender.FirstRq
	g.Loger <- [4]string{"GenFilter.ErrDaemon", "установнел хост отправителя", fmt.Sprintf("sendrHost: %s|limit:%d", g.Sender.HostRepiter, g.limit), "INFO"}
	g.cashSenders[g.Sender] = struct {}{}
	go func() {
		var i int
		for {
			<-g.ErrReq // ошибка ответа или другая, нам не ваажно, есть ошибки, переключаем по правилу
			if g.cfg.Mock {
				g.Loger <- [4]string{"GenFilter.ErrDaemon", g.Sender.HostRepiter, "IsMock:true (игнорируем ошибку)", "INFO"}
				continue
			}
			if g.globalErr > g.cfg.StopErr {
				g.Loger <- [4]string{"GenFilter.ErrDaemon", fmt.Sprintf("globalErrPredel:%d", g.globalErr), "достигли максимальное количество ошибок на всех хостах, завершение работы", "WARNING"}
				tp.ExitWithSecTimeout(1)
			}
			i++
			// смотрим квоту ошибок
			if i > g.Sender.StopErr {
				// надо останавливаться или менять отправителя
				g.globalErr++
				i = 0
				g.replaceSender()
				continue
			}
			g.Loger <- [4]string{"GenFilter.ErrDaemon", "resp.error.counter", fmt.Sprintf("зарегистрировано ошибок:%d|предел:%d", i, g.limit), "WARNING"}
		}
	}()
}
func (g *GenFilter) CallThirdParty(ipAddress, userAgent string) (result model.IPQSRow, err error) {
	if ipAddress == "" {
		err = fmt.Errorf("iP address IS EMPTY")
		return
	}
	result.Timestamp = time.Now()
	result.Uag = userAgent
	heandler := model.Handle{
		Time:        result.Timestamp,
		Send:        !g.cfg.Mock,
		RedirectUrl: fmt.Sprintf("https://ipqualityscore.com/api/json/ip/%s/%s", g.Sender.IpqsKey, ipAddress),
		Params:      fmt.Sprintf("allow_public_access_points=true&fast=false&lighter_penalties=true&mobile=false&strictness=1&user_agent=%s", strings.ReplaceAll(userAgent, " ", "%20")),
		Method:      "GET",
	}
	dat, err := json.Marshal(heandler)

	respRpc, err := RpcRequest(g.cfg.Service, string(dat), g.Sender.HostRepiter, g.Loger)
	if respRpc != nil {
		result.RespStatus = respRpc.RespStatus
		result.RespCode = respRpc.RespCode
		result.RespBody = respRpc.RespBody
	}
	//fmt.Printf("RpcRequest:%v\n", respRpc)
	//if resp.StatusCode != 200 {
	//	err = fmt.Errorf("wrong response code: %d (%s)", resp.StatusCode, resp.Status)
	//	return
	//}

	return
}
func (g *GenFilter) calc() {
	g.currentTimestamp = time.Now().Add(- time.Duration(g.cfg.ChDb.Timezone) * time.Hour)
	g.minusIntervalTimestamp = g.currentTimestamp.Add(- (g.Interval - 1*time.Second))
	g.filter = model.LogFilter{
		TimestampFrom: g.minusIntervalTimestamp,
		TimestampTo:   g.currentTimestamp,
		Source:        model.Source("redirect"),
	}
	g.ticker = time.Tick(g.Interval)
}
func isUpdateTariff(text string) bool {
	indexWarning := strings.Index(text, "\"success\":true") // много ворнингов в сентри по, уберем их оттуда
	if indexWarning == -1 {
		return true // если данное вхождение текста (i/o timeout) не найдено, значит это не таймаут а ошибка
	}
	return false
}
func (g *GenFilter) circle() {
	rows, err := g.chRepo.Get(&g.filter)
	if err != nil {
		g.Loger <- [4]string{"GenFilter.Select", fmt.Sprintf("get[%v|%v]", g.filter.TimestampFrom, g.filter.TimestampTo), fmt.Sprintf("%v", err), "ERROR"}
		time.Sleep(5 * time.Second)
		g.circle()
	}
	var i int
	if len(rows) > 0 {
		g.Loger <- [4]string{"GenFilter.Select", "rows", fmt.Sprintf("extract: %d", len(rows)), "INFO"}
		for _, v := range rows {
			// проверить есть ли в кеше
			ipKey := v.Ip.String()
			//keu := fmt.Sprintf("%s%s", ipKey, v.UserAgent.String)
			rowIpqs := g.db.TableIpAddress.Get(ipKey)
			if rowIpqs != nil {
				rowIpqs.RefererId = rowIpqs.Id
				//g.db.Sequences.SetCashDetect(ipKey)
				g.reporting.SetCashDetect(ipKey)
			} else {
				i++
				if i > g.limit {
					g.Loger <- [4]string{"circle.replaceSender", fmt.Sprintf("CallThirdParty[sender:%s]", g.Sender.HostRepiter), fmt.Sprintf("достигли лимита:%v", g.limit), "INFO"}
					g.replaceSender()
				}
				result, err := g.CallThirdParty(v.Ip.String(), v.UserAgent.String)
				if err != nil {
					g.Loger <- [4]string{"circle", fmt.Sprintf("CallThirdParty[ip_key:%s]", ipKey), fmt.Sprintf("внутренние ошибки [err:%v|resu:%v]", err, result), "ERROR"}
					// g.reporting.SetErr(ipKey, fmt.Errorf("внутренние ошибки [err:%v|body:%s]", err, result.RespBody))
					// если это внутренние ошибки, на будем их регистрировать по правилам конфигурации
					continue
				}
				g.reporting.SenderHost = g.Sender.HostRepiter
				if isUpdateTariff(result.RespBody) {
					g.Loger <- [4]string{"circle", fmt.Sprintf("CallThirdParty[ip_key:%s]", ipKey), fmt.Sprintf("IPQS Error [isUpdateTariff:true] [host:%s|body: %v", g.reporting.SenderHost, result.RespBody), "ERROR"}
					g.reporting.SetErr(ipKey, fmt.Errorf("request not valid (success true not detect)|body:%s", result.RespBody))
					g.ErrReq <- struct {}{}
					continue
				}
				g.globalErr = 0
				g.db.TableIpAddress.SetNew(ipKey, &result)
				g.reporting.SetOk(ipKey, result.RespBody)
				g.Loger <- [4]string{"circle", fmt.Sprintf("[OK.OK]CallThirdParty[ip_key:%s][uag:%s]", ipKey, v.UserAgent.String), fmt.Sprintf("%v", result), "RESPONSE"}
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		//g.db.TableIPQS.SaveBath(ipqsRows)
	} else {
		g.Loger <- [4]string{"GenFilter.Select", "rows", "нет строчек для обработки"}
	}

	<-g.ticker
	g.calc()
	g.circle()
}
