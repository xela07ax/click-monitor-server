package main

import (
	"encoding/json"
	"fmt"
	"github.com/niubaoshu/gotiny"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/db"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"github.com/xela07ax/toolsXela/chLogger"
	"github.com/xela07ax/toolsXela/tp"
	"os"
	"path/filepath"
	"time"
)

func newConfig(configDir string) *model.Config {
	// Открываем конфиг
	fi,err := tp.OpenReadFile(configDir)
	if err != nil {
		fmt.Printf("Ошибка при открытии конфигурации %s: %s\n",configDir,err)
		tp.ExitWithSecTimeout(1)
	}
	config := new(model.Config)
	err = json.Unmarshal(fi, config)
	if err != nil {
		fmt.Printf("Ошибка чтения JSON %s: %s\n",configDir,err)
		tp.ExitWithSecTimeout(1)
	}
	return config
}

func main() {
	fmt.Printf("ಠ┗(▀̿Ĺ̯▀̿ ̿)┓ \n")
	fmt.Printf("        Read databale LevelDb\n")
	fmt.Printf("      ٩◔̯◔۶\n")
	// Подготовим конфиг
	dir, err := tp.BinDir()
	tp.Fck(err)
	cfgPth := filepath.Join(dir, "config.json")
	fmt.Printf("config path: %s\n", cfgPth)
	cfg := newConfig(cfgPth)
	cfg.Path.BinDir = dir
	cfg.ChExit = make(chan os.Signal,2)

	logEr := chLogger.NewChLoger(&chLogger.Config{Dir: cfg.Path.Log})
	logEr.RunMinion()

	slowPoke := db.NewStore(cfg, logEr.ChInLog)
	slowPoke.RunMinion()

	// переместим ip адреса в свою таблицу, там и будем с ней работать
	all := slowPoke.TableIPQS.ReadAll()
	for k, v := range all {
		ipqsSaveIp := new(model.IPQSRow)
		gotiny.Unmarshal([]byte(v[1]), ipqsSaveIp)
		ipqsSaveIp.Ip = k
		ipqsSaveIp.Uag = v[0]
		slowPoke.TableIpAddress.SetNew(k, ipqsSaveIp)
	}
	fmt.Println("▌ █║ exiting after 1 second │║")
	time.Sleep(200*time.Millisecond)
}
