package clickhouse

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/ClickHouse/clickhouse-go"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"log"
	"time"
)

const (
	logsTable        = "redirect_logs"
	logsFields       = "timestamp, id, click_id, hit_id, is_unique, is_trafficback, trafficback_reason, source, ip, url, user_agent, redirect_url, merchant_id, network_id, affiliate_id, offer_id, campaign_id, promo_id, status, country_iso, region_iso, city, hardware_type"
	CurrentTimeStamp = "2006-01-02 15:04:05"
	LogsTimeStamp    = "2006-01-02_15:04:05"
)

var (
	filterIsNilErr = fmt.Errorf("filter is nil")
)

type LogRepo struct {
	Loger    chan<- [4]string
	chDb     *sql.DB
	timezone uint32
}

func NewLogRepo(cfg *model.Config, loger chan<- [4]string) LogRepo {
	// Подключение к KITTENHOUSE
	//khDb, err := sql.Open("kittenhouse", cfg.ChDb.KittenhouseDsn)
	//if err != nil {
	//	loger <- [4]string{"NewClientRepo", "Подключение KITTENHOUSE", fmt.Sprintf("Error: %v\n", err), "ERROR"}
	//	return nil
	//}
	// Подключение к CLICKHOUSE_DB
	var err error
	var chDb *sql.DB
	if cfg.ChDb.Mock {
		return LogRepo{
			Loger: loger,
			chDb:  chDb,
		}
	}
	if cfg.ChDb.Ssh != nil {
		loger <- [4]string{"NewClientRepo", "Подключение CLICKHOUSE", fmt.Sprintf("Use SSH: %s\n", cfg.ChDb.Ssh.ClickhouseDsn), "INFO"}
		tunnel := sshTunnel(cfg.ChDb.Ssh) // Initialize sshTunnel
		go tunnel.Start()     // Start the sshTunnel
		time.Sleep(500*time.Millisecond)
		// fmt.Println("clickhouse:", cfg.ChDb.ClickhouseDsn)
		chDb, err = sql.Open("clickhouse", cfg.ChDb.Ssh.ClickhouseDsn)
		loger <- [4]string{"NewClientRepo", "Подключение CLICKHOUSE", "OPENEd", "INFO"}

	} else {
		chDb, err = sql.Open("clickhouse", cfg.ChDb.ClickhouseDsn)
	}

	if err != nil {
		loger <- [4]string{"NewClientRepo", "Подключение CLICKHOUSE_DB", fmt.Sprintf("Error: %v\n", err), "ERROR"}
		panic(err)
	}
	loger <- [4]string{"NewClientRepo", "Подключение CLICKHOUSE", "Ping", "INFO"}
	if err := chDb.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Fatalf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		log.Fatal(err)
	}
	loger <- [4]string{"NewClientRepo", "Подключение CLICKHOUSE", "Ping[ok]", "INFO"}
	rows, err := chDb.Query(`select click_id from hoqu_fiat.fraud_log LIMIT 1`)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(500*time.Millisecond)
	for rows.Next() {
		var (
			col1 string
		)
		if err := rows.Scan(&col1); err != nil {
			log.Fatal(err)
		}
		loger <- [4]string{"NewClientRepo", "Тест подключения", fmt.Sprintf("Успешно[row]: col1=%s", col1), "INFO"}
	}
	loger <- [4]string{"NewClientRepo", "Подключение CLICKHOUSE_DB", fmt.Sprintf("Успешно"), "INFO"}

	rows.Close()
	time.Sleep(500*time.Millisecond)

	return LogRepo{
		Loger: loger,
		chDb:  chDb,
	}
}

//func (r *LogRepo) GetFrodRowsFromFile() map[string]model.IPQSRow {
//	r.Loger <- [4]string{"LogRepo", "GetFrodRowsFromFile", "froud_logs.json"}
//	b, _ := tp.OpenReadFile("froud_logs.json")
//	var resultRows []model.FraudLog
//	json.Unmarshal(b, &resultRows)
//	str := `api\/json\/ip\/\w+\/(.+)?\?` // нам нужен ip адрес
//	rex := regexp.MustCompile(str)
//	upd := make(map[string]model.IPQSRow)
//	var uagBad int
//	for _, row := range resultRows{
//		es := rex.FindStringSubmatch(row.RequestUrl)
//		if len(es) < 1 {
//			panic("whattt!")
//		}
//		if val, ok := upd[es[1]]; ok {
//			//do something here
//			dur := val.Timestamp.Sub(row.Timestamp)
//
//			if dur < 10 {
//				u, err := url.Parse(row.RequestUrl)
//				if err != nil {
//					panic(err)
//				}
//				q, err := url.ParseQuery(u.RawQuery)
//				if err != nil {
//					panic(err)
//				}
//				uagInRequst := q.Get("user_agent")
//				if len(uagInRequst) < 10 {
//					uagBad++
//					continue
//				}
//				upd[es[1]] = model.IPQSRow{
//					Timestamp: row.Timestamp,
//					Ip:        es[1],
//					Uag:       uagInRequst,
//					Request:   row.RequestUrl,
//					RpcFields: model.RpcFields{
//						RespStatus: "okay",
//						RespBody:   row.ResponseBody,
//					},
//				}
//				//fmt.Print(dur)
//			}
//
//		} else {
//			// проверим UserAgent-ы
//			u, err := url.Parse(row.RequestUrl)
//			if err != nil {
//				panic(err)
//			}
//			q, err := url.ParseQuery(u.RawQuery)
//			if err != nil {
//				panic(err)
//			}
//			uagInRequst := q.Get("user_agent")
//			if len(uagInRequst) < 10 {
//				uagBad++
//				//fmt.Println("continue")
//				continue
//				//panic(uagInRequst)
//			}
//			upd[es[1]] = model.IPQSRow{
//				Timestamp: row.Timestamp,
//				Ip:        es[1],
//				Uag:       uagInRequst,
//				Request:   row.RequestUrl,
//				RpcFields: model.RpcFields{
//					RespStatus: "okay",
//					RespBody:   row.ResponseBody,
//				},
//			}
//		}
//		//upd[es[1]] = model.FraudLog{
//		//	Timestamp:    time.Time{},
//		//	RequestUrl:   "",
//		//	ResponseBody: "",
//		//}
//	}
//	r.Loger <- [4]string{"LogRepo", "GetFrodRows.save", "froud_ipqs.json"}
//	f, _ := tp.CreateOpenFile("froud_ipqs.json")
//	b=nil
//	b, _ = json.Marshal(upd)
//	f.Write(b)
//	f.Close()
//	fmt.Println(uagBad)
//	fmt.Print("End\n")
//	return nil
//}

//func (r *LogRepo) GetFrodRows() []model.FraudLog {
//	query := "SELECT `timestamp`, request_url, response_body from fraud_log WHERE match(response_body, 'true')"
//	rows, err := r.chDb.Query(query)
//	if err != nil {
//		fmt.Printf("query: %s\n", query)
//		fmt.Printf("error: %v\n", err.Error())
//		panic(err)
//	}
//	r.Loger <- [4]string{"LogRepo", "GetFrodRows", "query.OK"}
//	var resultRows []model.FraudLog
//	var rowIndex int
//	for rows.Next() {
//		rowIndex++
//
//		log := model.FraudLog{}
//		err := rows.Scan(
//			&log.Timestamp,
//			&log.RequestUrl,
//			&log.ResponseBody,
//		)
//		if err != nil {
//			panic(err)
//		}
//		r.Loger <- [4]string{"LogRepo", "GetFrodRows", fmt.Sprintf("rows.Next(%d)", rowIndex)}
//		resultRows = append(resultRows, log)
//	}
//	r.Loger <- [4]string{"LogRepo", "GetFrodRows.save", "froud_logs.json"}
//	f, _ := tp.CreateOpenFile("froud_logs.json")
//	b, _ := json.Marshal(resultRows)
//	f.Write(b)
//	f.Close()
//	r.Loger <- [4]string{"LogRepo.GetFrodRows", "SelectRows", fmt.Sprint(len(resultRows))}
//	return resultRows
//}

func (r *LogRepo) Get(filter *model.LogFilter) (logs []*model.RedirectionLog, err error) {
	if filter == nil {
		err = filterIsNilErr
		return
	}
	fn := "LogRepo.Get.Query"
	alias := fmt.Sprintf("%s|%s|%d", filter.TimestampFrom.Format(LogsTimeStamp), filter.TimestampTo.Format(LogsTimeStamp), filter.Timezone)

	r.Loger <- [4]string{fn, alias, fmt.Sprintf("filter:%v", filter)}
	query := r.makeSqlSelect()
	where := filter.GetSqlWhere()
	order := filter.GetSqlOrderTimestamp()
	limit, err := filter.GetSqlLimit()
	if err != nil {
		return
	}
	queryFinal := fmt.Sprintf("%s %s %s %s;", query, where, order, limit)
	r.Loger <- [4]string{fn, alias, queryFinal}

	rows, err := r.chDb.Query(queryFinal)
	if err != nil {
		fmt.Printf("query: %s\n", query)
		fmt.Printf("where: %s\n", where)
		fmt.Printf("limit: %s\n", limit)
		fmt.Printf("error: %v\n", err.Error())
		return nil, fmt.Errorf("%s[%s]err:%v", fn, alias, err)
	}
	r.timezone = filter.Timezone
	return r.scanRows(rows)
}

func (r *LogRepo) makeSqlSelect() string {
	return fmt.Sprintf("SELECT %s FROM %s ", logsFields, logsTable)
}

func (r *LogRepo) makeSqlCount() string {
	return fmt.Sprintf("SELECT count(*) FROM %s", logsTable)
}

func (r *LogRepo) queryExec(query string, where string, limit string) (rows *sql.Rows, err error) {
	rows, err = r.chDb.Query(query, where, limit)
	if err != nil {
		return
	}

	return
}

func (r *LogRepo) scanRowsCount(rows *sql.Rows) (total int32, err error) {
	for rows.Next() {
		err = rows.Scan(&total)
		if err != nil {
			return
		}
	}

	return
}

type TimeStamp struct {
	dateTime *string
	timezone uint32
}

func (t TimeStamp) Scan(value interface{}) (err error) {
	switch v := value.(type) {
	case time.Time:
		*t.dateTime = value.(time.Time).Add(time.Duration(t.timezone) * time.Minute).Format(CurrentTimeStamp)
	default:
		return fmt.Errorf("invalid type for current_timestamp|your:%v", v)
	}
	return
}

func (t TimeStamp) Value() (driver.Value, error) {
	return t.dateTime, nil
}

func (r *LogRepo) scanRows(rows *sql.Rows) (logs []*model.RedirectionLog, err error) {
	for rows.Next() {
		log := model.RedirectionLog{}
		err = rows.Scan(
			TimeStamp{&log.Timestamp, r.timezone},
			&log.ID,
			&log.ClickID,
			&log.HitID,
			&log.IsUnique,
			&log.IsTrafficback,
			&log.TrafficbackReason,
			&log.Source,
			&log.Ip,
			&log.Url,
			&log.UserAgent,
			&log.RedirectUrl,
			&log.MerchantID,
			&log.NetworkID,
			&log.AffiliateID,
			&log.OfferID,
			&log.CampaignID,
			&log.PromoID,
			&log.Status,
			&log.CountryISO,
			&log.RegionISO,
			&log.City,
			&log.HardwareType,
		)
		if err != nil {
			return
		}

		logs = append(logs, &log)
	}

	return
}
