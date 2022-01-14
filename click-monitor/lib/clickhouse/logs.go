package clickhouse

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/xela07ax/rest-repiter/model"
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
	fmt.Println("clickhouse:", cfg.ChDb.ClickhouseDsn)
	chDb, err := sql.Open("clickhouse", cfg.ChDb.ClickhouseDsn)
	if err != nil {
		loger <- [4]string{"NewClientRepo", "Подключение CLICKHOUSE_DB", fmt.Sprintf("Error: %v\n", err), "ERROR"}
		panic(err)
	}

	return LogRepo{
		Loger: loger,
		chDb:  chDb,
	}
}

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
