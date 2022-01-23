package model

import (
	"os"
)

type Pathes struct {
	Log     string `json:"log"`
	Report  string `json:"report"`
	GenTime string `json:"gen_time"`
	BinDir  string
}

type Clickhouse struct {
	Timezone       int    `json:"timezone"`
	KittenhouseDsn string `json:"kittenhouse_dsn"`
	ClickhouseDsn  string `json:"clickhouse_dsn"`
}

type Report struct {
	DbPathUserAgent      string `json:"db_path_userAgent"`
	DbPath_IPaddress     string `json:"db_path_IPaddress"`
	DbPath_Ipqs          string `json:"db_path_Ipqs"`
	DbPath_sequence      string `json:"db_path_sequence"`
	DbSlowWriterDaley_Ms int    `json:"db_slowWriterDaley_Ms"`
	DbSequencesStart     int    `json:"db_sequencesStart"`
	TableUserAgent       string `json:"table_userAgent"`
	TableIPaddress       string `json:"table_IPaddress"`
	TableIpqs            string `json:"table_Ipqs"`
}

type Sender struct {
	FirstRq     int    `json:"first_rq"`
	ThinkRq     int    `json:"think_rq"`
	StopErr     int    `json:"stop_err"`
	Sleep       int    `json:"sleep"`
	IpqsKey     string `json:"ipqs_key"`
	HostRepiter string `json:"host_repiter"`
}

type Config struct {
	HttpServerPort int        `json:"HttpServerPort"`
	Interval       int        `json:"interval"`
	StopErr        int        `json:"stop_err"`
	Mock           bool       `json:"mock"`
	UrlPostback    string     `json:"url_postback"`
	Service        string     `json:"service"`
	Senders        []*Sender  `json:"senders"`
	Path           Pathes     `json:"path"`
	ChDb           Clickhouse `json:"clickhouse"`
	Reporting      Report     `json:"reporting"`
	// В конфиге не проставлять!
	KeySession []byte
	ChExit     chan os.Signal `json:"_"`
}
