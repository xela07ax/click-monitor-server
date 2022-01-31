package db

import (
	"fmt"
	"github.com/niubaoshu/gotiny"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"github.com/xela07ax/toolsXela/encod"
	"github.com/xela07ax/toolsXela/tp"
	"strings"
)

type IpAddressTable struct {
	name             string
	subName          string
	sequences        SequencesGen
	config           *model.Config
	dbDIDtoClientUid *leveldb.DB
	db               *leveldb.DB
	loger            chan<- [4]string
}

func (ut *IpAddressTable) ReadAll() []model.IPQSRow {
	resultIpuag := make([]model.IPQSRow, 0, 10000)
	iter := ut.db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		normal := model.IPQSRow{}
		gotiny.Unmarshal(iter.Value(), &normal)
		normal.SenderIp = string(key)
		resultIpuag = append(resultIpuag, normal)
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		panic(err)
	}
	return resultIpuag
}

func (ut *IpAddressTable) SetNew(key string, ipqRow *model.IPQSRow) {
	ipqRow.Id = ut.sequences.GetNewUidTable(ut.sequences.Tables.IP)
	if err := ut.db.Put([]byte(key), gotiny.Marshal(ipqRow), nil); err != nil {
		ut.loger <- [4]string{ut.name, "Put", fmt.Sprintf("Не удалось записатьв таблицу IPQS| ERTX:%v", err), "1"}
		tp.ExitWithSecTimeout(1)
	}
}
func (ut *IpAddressTable) Del(key string) {
	err := ut.db.Delete([]byte(key), nil)
	if err != nil {
		fn := fmt.Sprintf("=>%s", strings.Join([]string{ut.subName, fmt.Sprintf("Delete UserDetail %s", key)}, "=>"))
		ut.loger <- [4]string{ut.name, "nil", fmt.Sprintf("%s | Не удалось Delete пользователя по Id из SlowБД | %v", fn, err), "ERROR"}
		tp.ExitWithSecTimeout(1)
	}
}

func (ut *IpAddressTable) Get(id string) *model.IPQSRow {
	userBt, err := ut.db.Get([]byte(id), nil)
	if err == leveldb.ErrNotFound {
		// пусто
		return nil
	} else if err != nil {
		fn := fmt.Sprintf("=>%s", strings.Join([]string{ut.subName, fmt.Sprintf("Get UserDetail %s", id)}, "=>"))
		ut.loger <- [4]string{ut.name, "nil", fmt.Sprintf("%s | Не удалось считать пользователя по Id из SlowБД | %v", fn, err), "ERROR"}
		tp.ExitWithSecTimeout(1)
	}
	// Данные найдены
	var row model.IPQSRow
	gotiny.Unmarshal(userBt, &row)
	return &row
}

//func (dt *IpAddressTable) SaveDialog(dialog model.IPQSRow) {
//	fn := fmt.Sprintf("=>%s",strings.Join([]string{dt.subName,"Save IPQSRow"},"=>"))
//	if err := dt.db.Put(encod.IntToBytes(dialog.DID),gotiny.Marshal(&dialog),nil); err != nil {
//		dt.loger<- [4]string{dt.name,"nil",fmt.Sprintf("%s | COM:Не удалось записать в IpAddressTable | ERTX:%v",fn,err),"1"}
//		dt.config.ExitProgramErr()
//	}
//}

func (dt *IpAddressTable) SaveBath(ipqs []model.IPQSRow) {
	fn := fmt.Sprintf("=>%s", strings.Join([]string{dt.subName, "Save bath IPQSRow"}, "=>"))
	batch := new(leveldb.Batch)
	for _, row := range ipqs {
		id := dt.sequences.GetNewUidTable(dt.sequences.Tables.IPQS)
		batch.Put(encod.IntToBytes(id), gotiny.Marshal(&row))
	}
	err := dt.db.Write(batch, nil)
	if err != nil {
		dt.loger <- [4]string{dt.name, "nil", fmt.Sprintf("%s | COM:Не удалось записать в IpAddressTable Batch-ем | ERTX:%v", fn, err), "1"}
		tp.ExitWithSecTimeout(1)
	}
}
