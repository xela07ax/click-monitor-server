package reporter

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"github.com/xela07ax/toolsXela/tp"
	"path/filepath"
	"strings"
	"time"
)

type Reporting struct {
	SenderHost   string
	cfg   *model.Config
	Loger chan<- [4]string
}

func NewReporting(cfg *model.Config, loger chan<- [4]string) *Reporting {
	err := tp.CheckMkdir(filepath.Join(cfg.Path.BinDir, cfg.Path.Report))
	if err != nil {
		panic(err)
	}
	return &Reporting{cfg: cfg, Loger: loger}
}

func (r *Reporting) SetErr(key string, errSv error) {
	b, pthFile := r.openFile(fmt.Sprintf("%s_resp_err_%s", getDay(), cleanHost(r.SenderHost)))
	r.writeFile(pthFile, []byte(fmt.Sprintf("[%v]%s\n%s", tp.Getime(), fmt.Sprintf("%s║▌%v", key, errSv), b)))
}

func cleanHost(host string) string {
	host = strings.ReplaceAll(host,"http://","")
	host = strings.ReplaceAll(host,".","")
	host = strings.ReplaceAll(host,":","")
	host = strings.ReplaceAll(host,"/rpc","")
	return host
}
func getDay() string  {
	return time.Now().Format("20060102")
}

func (r *Reporting) SetOk(key, body string) {
	tp.Getime()
	b, pthFile := r.openFile(fmt.Sprintf("%s_resp_ok_%s", getDay(), cleanHost(r.SenderHost)))
	r.writeFile(pthFile, []byte(fmt.Sprintf("[%v]%s\n%s", tp.Getime(), fmt.Sprintf("%s║▌%v", key, body), b)))
}

func (r *Reporting) openFile(name string) (dataFile []byte, path string) {
	path = filepath.Join(r.cfg.Path.BinDir, r.cfg.Path.Report, fmt.Sprintf("%s.json", name))
	dataFile, _ = tp.OpenReadFile(filepath.Join(r.cfg.Path.BinDir, r.cfg.Path.Report, fmt.Sprintf("%s.json", name)))
	return
}
func (r *Reporting) writeFile(pthFile string, dataFile []byte) {
	f, err := tp.CreateOpenFile(pthFile)
	if err != nil {
		r.Loger <- [4]string{"writeFile", "tp.CreateOpenFile", fmt.Sprintf(" Не удалось считать %s.json пользователя| %v", pthFile, err), "ERROR"}
		tp.ExitWithSecTimeout(1)
	}
	_, err = f.Write(dataFile)
	if err != nil {
		r.Loger <- [4]string{"writeFile", "f.Write", fmt.Sprintf(" Не удалось считать %s.json пользователя| %v", pthFile, err), "ERROR"}
		tp.ExitWithSecTimeout(1)
	}
	err = f.Close()
	if err != nil {
		r.Loger <- [4]string{"writeFile", "f.Close", fmt.Sprintf(" Не удалось считать %s.json пользователя| %v", pthFile, err), "ERROR"}
		tp.ExitWithSecTimeout(1)
	}

	return
}
func (r *Reporting) SetCashDetect(key string) {
	b, pthFile := r.openFile("cash_detect")
	var counter int
	report := make(map[string]int)
	if len(b) > 0 {
		err := json.Unmarshal(b, &report)
		if err != nil {
			r.Loger <- [4]string{"SetCashDetect", "ioutil.ReadAll", fmt.Sprintf(" Не удалось считать cash_detect.json ioutil.ReadAll | %v", err), "ERROR"}
			tp.ExitWithSecTimeout(1)
		}
		if num, ok := report[key]; ok {
			counter = num
		}
	}
	counter++
	report[key] = counter
	dat, err := json.Marshal(report)
	if err != nil {
		r.Loger <- [4]string{"SetCashDetect", "json.Marshal", fmt.Sprintf(" Не удалось считать cash_detect.json json.Marshal | %v", err), "ERROR"}
		tp.ExitWithSecTimeout(1)
	}
	r.writeFile(pthFile, dat)
}
