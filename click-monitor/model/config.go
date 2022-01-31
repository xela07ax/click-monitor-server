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

type Ssh struct {
	ServerHost     string `json:"sshServerHost"`
	ServerPort     int    `json:"sshServerPort"`
	UserName       string `json:"sshUserName"`
	PrivateKeyFile string `json:"sshPrivateKeyFile"`
	LocalHost      string `json:"sshLocalHost"`
	LocalPort      int    `json:"sshLocalPort"`
	RemoteHost     string `json:"sshRemoteHost"`
	RemotePort     int    `json:"sshRemotePort"`
	ClickhouseDsn  string `json:"clickhouse_dsn"`
}

type Clickhouse struct {
	Mock           bool   `json:"mock"`
	Ssh            *Ssh   `json:"ssh"`
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
	Name        string `json:"name"`
	FirstRq     int    `json:"first_rq"`
	ThinkRq     int    `json:"think_rq"`
	StopErr     int    `json:"stop_err"`
	ErrDetect   bool   `json:"err_detect"`
	Sleep       int    `json:"sleep"`
	IpqsKey     string `json:"ipqs_key"`
	HostRepiter string `json:"host_repiter"`
}
type Mode struct {
	Counter           bool `json:"counter"`
	Updater           bool `json:"updater"`
	ClickMonitor      bool `json:"click_monitor"`
	ClickMonitorAsync bool `json:"click_monitor_async"`
}

type Config struct {
	HttpServerPort int        `json:"HttpServerPort"`
	Interval       int        `json:"interval"`
	StopErr        int        `json:"stop_err"`
	Mock           bool       `json:"mock"`
	CashLenth      int        `json:"cash_lenth"`
	UrlPostback    string     `json:"url_postback"`
	Service        string     `json:"service"`
	Senders        []*Sender  `json:"senders"`
	ModeStart      Mode       `json:"mode_start"`
	Path           Pathes     `json:"path"`
	ChDb           Clickhouse `json:"clickhouse"`
	Reporting      Report     `json:"reporting"`
	// В конфиге не проставлять!
	KeySession []byte
	ChExit     chan os.Signal `json:"_"`
}
