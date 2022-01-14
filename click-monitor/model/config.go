package model

import (
	"os"
	"time"
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

type Send struct {
	Send        bool   `json:"send"`
	Ip          string `json:"ip"`
	HostRepiter string `json:"host_repiter"`
	Service     string `json:"service"`
}

type Config struct {
	HttpServerPort int        // Порт программы
	Path           Pathes     `json:"path"`
	ChDb           Clickhouse `json:"clickhouse"`
	StartDate      time.Time  `json:"start_date"`
	Interval       int        `json:"interval"`
	Sender         Send       `json:"sender"`
	IpqsKey        string     `json:"ipqs_key"`
	Reporting      Report     `json:"reporting"`
	// В конфиге не проставлять!
	KeySession []byte
	ChExit     chan os.Signal `json:"_"`
}

//
//func (cfg *Config) ExitProgramErr() {
//	cfg.exitProgram(1)
//}
//func (cfg *Config) ExitProgramNorm() {
//	cfg.exitProgram(0)
//}
//func (cfg *Config) exitProgram(status int) {
//	// Го любит завершать свою работу раньше чем сделать все завершающие операции, но все же он остается очень быстрым если убрать задержку и немного подождать
//	// 0 - norm
//	// 1 - error
//	fmt.Println("Завершение работы программы, ускоряем выдачу логов")
//	// time.Sleep(1*time.Second)
//	// Внимание!, обработка статусов выхода временно недоступно, выход всегда безошибочный*
//	//os.Exit(status)
//	// Отправляем сигнал завершения в функцию безопасного выхода "Daemon"
//	cfg.ChExit <- syscall.SIGSTOP
//}
