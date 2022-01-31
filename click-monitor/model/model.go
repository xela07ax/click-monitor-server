package model

import (
	"database/sql"
	"net"
	"time"
)

type Body struct {
	Request string `json:"request"`
}

type Micro struct {
	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
}

type RpcModel struct {
	Micro
	Body
}

type RpcFields struct {
	RespStatus string
	RespCode   int
	RespBody   string
}

type Pino struct {
	Message string
	Err     error
	Traffic []byte
	Door    chan Pino
}

type Handle struct {
	Time                                    time.Time
	Send                                    bool
	RedirectUrl, Params, Method, Body, Type string
}

type IpqsPrepareData struct {
	Ip string
	UserAgent string
}

type FraudLog struct {
	Timestamp time.Time
	RequestUrl string
	ResponseBody string
}

type RedirectionLog struct {
	Timestamp         string
	ID                string
	ClickID           sql.NullString
	HitID             sql.NullString
	IsUnique          sql.NullBool
	IsTrafficback     sql.NullBool
	TrafficbackReason sql.NullString
	Source            sql.NullString
	Ip                net.IP
	Url               sql.NullString
	UserAgent         sql.NullString
	RedirectUrl       sql.NullString
	MerchantID        sql.NullInt32
	NetworkID         sql.NullInt32
	AffiliateID       sql.NullInt32
	OfferID           sql.NullInt32
	CampaignID        sql.NullInt32
	PromoID           sql.NullInt32
	Status            sql.NullString
	CountryISO        sql.NullString
	RegionISO         sql.NullString
	City              sql.NullString
	HardwareType      sql.NullString
}

type IPQSRow struct {
	Id        int
	RefererId int
	Timestamp time.Time
	Ip        string
	Uag       string
	SenderIp  string
	Request   string
	RpcFields
}

type Requests struct {
	Key_ipUag   string
	Val_IPQSRow string
}

var CashDetect map[string]int = make(map[string]int)
