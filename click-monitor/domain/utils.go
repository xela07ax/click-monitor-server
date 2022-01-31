package domain

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"github.com/xela07ax/toolsXela/tp"
	"log"
)

func (g *GenFilter) UpdateDatabase() {
	upd := make(map[string]*model.IPQSRow)
	add := make(map[string]*model.IPQSRow)
	del := make(map[string]interface{})
	var b []byte
	var err error
	b, err = tp.OpenReadFile("upd.json")
	if err != nil {
		log.Fatalf("upd.json[open.ERR]:%s", err)
	}
	err = json.Unmarshal(b, &upd)
	if err != nil {
		log.Fatalf("upd.json[unmarshal.ERR]:%s", err)
	}
	b = nil

	b, err = tp.OpenReadFile("add.json")
	if err != nil {
		log.Fatalf("add.json[open.ERR]:%s", err)
	}
	err = json.Unmarshal(b, &add)
	if err != nil {
		log.Fatalf("add.json[unmarshal.ERR]:%s", err)
	}
	b = nil

	b, err = tp.OpenReadFile("del.json")
	if err != nil {
		log.Fatalf("del.json[open.ERR]:%s", err)
	}
	err = json.Unmarshal(b, &del)
	if err != nil {
		log.Fatalf("del.json[unmarshal.ERR]:%s", err)
	}
	b = nil

	var deleted int
	for k, _ := range del {
		if g.DelDatabase(k) {
			deleted++
		}
	}

	var updated int
	for k, v := range upd {
		g.UpdDatabase(k, v)
		updated++
	}

	var added int
	var fosed int
	for k, v := range add {
		if g.AddDatabase(k, v) {
			added++
		} else {
			fosed++
		}
	}
	g.Loger <- [4]string{"del", "LENGTH", fmt.Sprint(len(del))}
	g.Loger <- [4]string{"upd", "LENGTH", fmt.Sprint(len(upd))}
	g.Loger <- [4]string{"add", "LENGTH", fmt.Sprint(len(add))}
	g.Loger <- [4]string{"del", "report", fmt.Sprintf("deleted:%d", deleted)}
	g.Loger <- [4]string{"upd", "report", fmt.Sprintf("updated:%d", updated)}
	g.Loger <- [4]string{"add", "report", fmt.Sprintf("added:%d|fosed:%d", added, fosed)}


	fmt.Println(" Gooodby!")
	tp.ExitWithSecTimeout(0)
}


func (g *GenFilter) DelDatabase(ipKey string) bool {
	rowIpqs := g.db.TableIpAddress.Get(ipKey)
	if rowIpqs != nil {
		g.db.TableIpAddress.Del(ipKey)
		return true
	}
	return false
}

func (g *GenFilter) UpdDatabase(ipKey string, ipqsModel *model.IPQSRow) {
	g.db.TableIpAddress.SetNew(ipKey, ipqsModel)
}

func (g *GenFilter) AddDatabase(ipKey string, ipqsModel *model.IPQSRow) bool {
	rowIpqs := g.db.TableIpAddress.Get(ipKey)
	if rowIpqs != nil {
		return false
	}
	g.db.TableIpAddress.SetNew(ipKey, ipqsModel)
	return true
}
