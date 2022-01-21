package domain

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/clickhouse"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/db"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/reporter"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
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
	Loger                  chan<- [4]string
}

func New_GenFilter_ChDbMonitor(cfg *model.Config, loger chan<- [4]string) *GenFilter {
	slowPoke := db.NewStore(cfg, loger)
	slowPoke.RunMinion()

	gen := &GenFilter{db: slowPoke, reporting: reporter.NewReporting(cfg, loger), Interval: time.Duration(cfg.Interval) * time.Second, Loger: loger, chRepo: clickhouse.NewLogRepo(cfg, loger), cfg: cfg}
	gen.calc()

	go gen.circle()
	return gen
}
func (g *GenFilter) CallThirdParty(ipAddress, userAgent string) (result model.IPQSRow, err error) {
	if ipAddress == "" {
		err = fmt.Errorf("iP address IS EMPTY")
		return
	}
	result.Timestamp = time.Now()
	result.SenderIp = g.cfg.Sender.Ip
	result.Uag = userAgent
	heandler := model.Handle{
		Time:        result.Timestamp,
		Send:        g.cfg.Sender.Send,
		RedirectUrl: fmt.Sprintf("https://ipqualityscore.com/api/json/ip/%s/%s", g.cfg.IpqsKey, ipAddress),
		Params:      fmt.Sprintf("allow_public_access_points=true&fast=false&lighter_penalties=true&mobile=false&strictness=1&user_agent=%s", strings.ReplaceAll(userAgent, " ", "%20")),
		Method:      "GET",
	}
	dat, err := json.Marshal(heandler)

	respRpc, err := RpcRequest(g.cfg.Sender.Service, string(dat), g.cfg.Sender.HostRepiter, g.Loger)
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
	g.currentTimestamp = time.Now().Add(- time.Duration(g.cfg.ChDb.Timezone)*time.Hour)
	g.minusIntervalTimestamp = g.currentTimestamp.Add(- (g.Interval-1*time.Second))
	g.filter = model.LogFilter{
		TimestampFrom: g.minusIntervalTimestamp,
		TimestampTo:   g.currentTimestamp,
		Source: model.Source("redirect"),
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
		g.Loger <- [4]string{"GenFilter", fmt.Sprintf("get[%v|%v]", g.filter.TimestampFrom, g.filter.TimestampTo), fmt.Sprintf("%v", err), "ERROR"}
		time.Sleep(5 * time.Second)
		g.circle()
	}
	if len(rows) > 0 {
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
				result, err := g.CallThirdParty(v.Ip.String(), v.UserAgent.String)
				if err != nil {
					g.Loger <- [4]string{"circle", fmt.Sprintf("CallThirdParty[ip_key:%s]", ipKey), fmt.Sprintf("err:%v|resu:%v", err, result), "ERROR"}
					g.reporting.SetErr(ipKey, fmt.Errorf("err:%v|body:%s", err, result.RespBody))
					continue
				}
				if isUpdateTariff(result.RespBody) {
					g.Loger <- [4]string{"circle", fmt.Sprintf("CallThirdParty[ip_key:%s]", ipKey), fmt.Sprintf("err: (обычно такое)Please upgrade to increase your request quota.|resu:%v", result), "ERROR"}
					g.reporting.SetErr(ipKey, fmt.Errorf("request not valid (success true not detect)|body:%s", result.RespBody))
					continue
				}
				g.db.TableIpAddress.SetNew(ipKey, &result)
				//g.db.TableIPQS.SetNew(keu, result)
				g.reporting.SetOk(ipKey, result.RespBody)
				g.Loger <- [4]string{"circle", fmt.Sprintf("[OK.OK]CallThirdParty[ip_key:%s][uag:%s]", ipKey, v.UserAgent.String), fmt.Sprintf("%v", result), "RESPONSE"}
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		//g.db.TableIPQS.SaveBath(ipqsRows)
	} else {
		g.Loger <- [4]string{"GenFilter", fmt.Sprintf("rows[%v|%v]", g.filter.TimestampFrom, g.filter.TimestampTo), fmt.Sprintf("%v", err), "NULL"}
	}

	<-g.ticker
	g.calc()
	g.circle()
}
