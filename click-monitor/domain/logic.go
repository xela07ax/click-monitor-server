package domain

import (
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"strings"
	"time"
)

func (g *GenFilter) checkIsCashAndSend(log *model.RedirectionLog, requestWorkers chan<- model.IpqsPrepareData) {
	// проверить есть ли в кеше
	ipKey := log.Ip.String()
	rowIpqs := g.db.TableIpAddress.Get(ipKey)
	if rowIpqs != nil {
		g.reporting.SetCashDetect(ipKey)
		return // найден в кеше, запишем в репорт и все!
	}
	if _, ok := g.sended[ipKey]; ok {
		return // такой запрос уже отправляли, откажем в отправке
	}
	// если не отправляли, закешируем в миникеше
	g.sended[ipKey] = struct{}{}
	// в нашей базе такого ip нету, надо запросить инфу
	g.Loger <- [4]string{"GenFilter.toWorker", "запрашиваем информацию", fmt.Sprintf("помещаем в очередь на отправку:%s|cash:%d", ipKey, len(requestWorkers))}
	requestWorkers <- model.IpqsPrepareData{
		Ip:        ipKey,
		UserAgent: log.UserAgent.String,
	}
}

func (g *GenFilter) RequestDaemonWriter(respOkCh <-chan model.IPQSRow, respErCh <-chan [2]string) {
	for {
		select {
		case okResp := <-respOkCh:
			g.db.TableIpAddress.SetNew(okResp.Ip, &okResp)
			g.reporting.SetOk(okResp.Ip, okResp.RpcFields.RespBody, okResp.SenderIp)
		case errResp := <-respErCh:
			g.reporting.SetErr(errResp[0], fmt.Errorf("%s", errResp[1]))
		}
	}
}

func (g *GenFilter) DaemonSqlMaker(inputTimes chan [2]time.Time, requestWorkers chan<- model.IpqsPrepareData) bool {
	for {
		tims := <-inputTimes
		g.Loger <- [4]string{"SqlMaker", fmt.Sprintf("newTick[from:%v|to:%v]", tims[1], tims[0]), "компилируем sql"}
		// даты для sql получили, соберем фильтр
		filter := model.LogFilter{
			TimestampFrom: tims[1],
			TimestampTo:   tims[0],
			Source:        model.Source("redirect"),
		}
		// делаем запрос и получаем строки из базы
		rows, err := g.chRepo.Get(&filter)
		if err != nil {
			g.Loger <- [4]string{"SqlMaker", fmt.Sprintf("получение строк из базы"), fmt.Sprintf("errSelect:%v| wait 10 second and continue", err), "ERROR"}
			time.Sleep(10 * time.Second)
			continue
		}

		g.Loger <- [4]string{"SqlMaker", fmt.Sprintf("получение строк из базы"), fmt.Sprintf("строк в работу:%d", len(rows))}
		for _, row := range rows {
			// отправим на
			g.checkIsCashAndSend(row, requestWorkers)
			// если воркеров не будет, то мы здесь зависнем пока не начнется чтение с заполненной очереди
		}
	}
}

func Validater(bodyText string) bool {
	indexWarning := strings.Index(bodyText, "\"success\":true")
	if indexWarning == -1 {
		return true // если заданное вхождение не найдено
	}
	return false
}
