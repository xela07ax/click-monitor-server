package model

import (
	"fmt"
	"html"
	"strings"
	"time"
)

const (
	TrafficbackAll   Trafficback = 0
	TrafficbackFalse Trafficback = 1
	TrafficbackTrue  Trafficback = 2

	SourceBot      Source = "bot"
	SourceGPR      Source = "google_parallel_redirect"
	SourceRedirect Source = "redirect"
	SourceClickAPI Source = "api"

	StatusSuccess Status = "success"
	StatusError   Status = "error"

	PageSizeMax     int32    = 100
	PageSizeDefault PageSize = 25
)

var (
	InvalidSourceErr      = fmt.Errorf("invalid source")
	InvalidTrafficbackErr = fmt.Errorf("invalid trafficback")
	InvalidStatusErr      = fmt.Errorf("invalid status")
	InvalidPageSize       = fmt.Errorf("invalid page size")
)

type Trafficback int32

func (s *Trafficback) IsValid() (err error) {
	switch *s {
	case TrafficbackAll, TrafficbackTrue, TrafficbackFalse:
	default:
		err = InvalidTrafficbackErr
		return
	}

	return
}

type Source string

func (s *Source) IsValid() (err error) {

	switch *s {
	case SourceBot, SourceGPR, SourceRedirect, SourceClickAPI:
	default:
		err = InvalidSourceErr
		return
	}

	return
}

type Status string

func (s *Status) IsValid() (err error) {
	switch *s {
	case StatusSuccess, StatusError:
	default:
		err = InvalidStatusErr
		return
	}

	return
}

type PageSize int32

func (s *PageSize) IsValid() (err error) {
	if int32(*s) > PageSizeMax || int32(*s) < 0 {
		err = InvalidPageSize
		return
	}

	return
}

func arrayToString(a []uint32, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

type LogFilter struct {
	Id            string
	Ip            string
	ClickID       string
	NetworkID     uint32
	OfferID       uint32
	CampaignID    uint32
	PromoID       uint32
	Affiliate     []uint32
	Merchant      []uint32
	Source        Source
	Status        Status
	Trafficback   Trafficback
	Url           string
	RedirectUrl   string
	TimestampFrom time.Time
	TimestampTo   time.Time
	Timezone      uint32

	RowsOffset int32
	PageSize   PageSize

	params []string
}

func (f *LogFilter) GetSqlLimit() (limitSql string, err error) {
	if int32(f.PageSize) == 0 {
		f.PageSize = PageSizeDefault
	}
	offset := f.RowsOffset
	limit := int32(f.PageSize)

	limitSql = fmt.Sprintf("LIMIT %v,%v", offset, limit)
	return
}

func (f *LogFilter) GetSqlWhere() (where string) {
	f.addWhereParams()
	f.params = append(f.params, "network_id != 43") // Demo network
	f.params = append(f.params, "not match(user_agent, '^\\d')")

	if len(f.params) == 0 {
		return
	}

	sqlParams := strings.Join(f.params, " AND ")
	where = fmt.Sprintf("WHERE %s", sqlParams)

	return
}
func (f *LogFilter) GetSqlOrderTimestamp() (order string) {
	order = "ORDER BY timestamp DESC"
	return
}

func (f *LogFilter) GetSqlCompiledWhere() (where string) {
	if len(f.params) == 0 {
		return
	}

	sqlParams := strings.Join(f.params, " AND ")
	where = fmt.Sprintf("WHERE %s", sqlParams)

	return
}

func (f *LogFilter) addWhereParams() {
	// Выполняется последовательно, так как разбитие на несколько горутин только замедлит работу
	f.addId()
	f.addIp()
	f.addClickID()
	f.addNetworkID()
	f.addOfferID()
	f.addCampaignID()
	f.addPromoID()
	f.addAffiliateID()
	f.addMerchantID()
	f.addSource()
	f.addStatus()
	f.addIsTrafficback()
	f.addUrl()
	f.addRedirectUrl()
	f.addTimestamp()
}

func (f *LogFilter) addId() {
	if f.Id != "" {
		f.params = append(f.params, fmt.Sprintf("id = '%s'", html.EscapeString(f.Id)))
	}
}
func (f *LogFilter) addIp() {
	if f.Ip != "" {
		f.params = append(f.params, fmt.Sprintf("ip = '%s'", html.EscapeString(f.Ip)))
	}
}

func (f *LogFilter) addClickID() {
	if f.ClickID != "" {
		f.params = append(f.params, fmt.Sprintf("click_id = '%s'", html.EscapeString(f.ClickID)))
	}
}

func (f *LogFilter) addNetworkID() {
	if f.NetworkID != 0 {
		f.params = append(f.params, fmt.Sprintf("network_id = %v", f.NetworkID))
	}
}

func (f *LogFilter) addOfferID() {
	if f.OfferID != 0 {
		f.params = append(f.params, fmt.Sprintf("offer_id = %v", f.OfferID))
	}
}

func (f *LogFilter) addCampaignID() {
	if f.CampaignID != 0 {
		f.params = append(f.params, fmt.Sprintf("campaign_id = %v", f.CampaignID))
	}
}

func (f *LogFilter) addPromoID() {
	if f.PromoID != 0 {
		f.params = append(f.params, fmt.Sprintf("promo_id = %v", f.PromoID))
	}
}

func (f *LogFilter) addAffiliateID() {
	if len(f.Affiliate) > 0 {
		f.params = append(f.params, fmt.Sprintf("affiliate_id in (%v)", arrayToString(f.Affiliate, ",")))
	}
}

func (f *LogFilter) addMerchantID() {
	if len(f.Merchant) > 0 {
		f.params = append(f.params, fmt.Sprintf("merchant_id in %v", arrayToString(f.Merchant, ",")))
	}
}

func (f *LogFilter) addSource() {
	if f.Source != "" {
		f.params = append(f.params, fmt.Sprintf("source = '%s'", f.Source))
	}
}

func (f *LogFilter) addStatus() {
	if f.Status != "" {
		f.params = append(f.params, fmt.Sprintf("status = '%s'", f.Status))
	}
}

func (f *LogFilter) addIsTrafficback() {
	if f.Trafficback == TrafficbackAll {
		return
	}

	var isTrafficback bool
	if f.Trafficback == TrafficbackTrue {
		isTrafficback = true
	}

	if f.Trafficback == TrafficbackFalse {
		isTrafficback = false
	}

	f.params = append(f.params, fmt.Sprintf("is_trafficback = %v", isTrafficback))
}

func (f *LogFilter) addUrl() {
	if f.Url != "" {
		f.params = append(f.params, fmt.Sprintf("ilike(url, '%%%s%%')", html.EscapeString(f.Url)))
	}
}

func (f *LogFilter) addRedirectUrl() {
	if f.RedirectUrl != "" {
		f.params = append(f.params, fmt.Sprintf("ilike(redirect_url, '%%%s%%')", html.EscapeString(f.RedirectUrl)))
	}
}

func (f *LogFilter) addTimestamp() {
	nilTime := time.Time{}
	if f.TimestampFrom == nilTime && f.TimestampTo == nilTime {
		return
	}

	if f.TimestampTo == nilTime {
		f.params = append(f.params, fmt.Sprintf("timestamp > '%s'", f.TimestampFrom.Format("2006-01-02 15:04:05")))
		return
	}

	if f.TimestampFrom == nilTime {
		f.params = append(f.params, fmt.Sprintf("timestamp < '%s'", f.TimestampTo.Format("2006-01-02 15:04:05")))
		return
	}

	f.params = append(f.params, fmt.Sprintf("timestamp between '%s' and '%s'", f.TimestampFrom.Format("2006-01-02 15:04:05"), f.TimestampTo.Format("2006-01-02 15:04:05")))
}

