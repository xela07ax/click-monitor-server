package db

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/xela07ax/rest-repiter/model"
	"github.com/xela07ax/toolsXela/encod"
	"github.com/xela07ax/toolsXela/tp"
	"strings"
)

type NameTables struct {
	IP   []byte
	UAG  []byte
	IPQS []byte
}

type SequencesGen struct {
	name    string
	subName string
	Tables  NameTables
	config  *model.Config
	db      *leveldb.DB
	loger   chan<- [4]string
}

func (sc *SequencesGen) sequenceInit() {
	subName := fmt.Sprintf("=>%s", strings.Join([]string{sc.subName, "Initialize"}, "=>"))
	snBt := encod.IntToBytes(sc.config.Reporting.DbSequencesStart)
	for _, tableNam := range [][]byte{sc.Tables.IP, sc.Tables.UAG, sc.Tables.IPQS} {
		sc.loger <- [4]string{sc.name, "nil", fmt.Sprintf("%s | COM:Проверяем последовательность таблицы %s ", subName, tableNam)}
		val, err := sc.db.Get(tableNam, nil)
		if err == leveldb.ErrNotFound {
			sc.loger <- [4]string{sc.name, "nil", fmt.Sprintf("%s | COM:Последовательность %s пуста, проставляем sn: [%v]", subName, tableNam, snBt)}
			if err := sc.db.Put(tableNam, snBt, nil); err != nil {
				sc.loger <- [4]string{sc.name, "nil", fmt.Sprintf("%s | COM:Не удалось записать ключ в БД | Key:%s | ERTX:%v", subName, tableNam, err), "1"}
				tp.ExitWithSecTimeout(1)
			}
		} else if err != nil {
			sc.loger <- [4]string{sc.name, "nil", fmt.Sprintf("%s | COM:Не удалось считать последовательность таблицы %s| ERTX:%v", subName, tableNam, err), "1"}
			tp.ExitWithSecTimeout(1)
		} else {
			// Данные найдены
			sc.loger <- [4]string{sc.name, "nil", fmt.Sprintf("%s | COM:Значение таблицы %s | DT:%d", subName, tableNam, encod.BytesToInt(val))}
		}
	}
}
func (sc *SequencesGen) SetCashDetect(keyIpUag string) {
	id := sc.GetCashDetect(keyIpUag)
	if err := sc.db.Put([]byte(keyIpUag), encod.IntToBytes(id), nil); err != nil {
		sc.loger <- [4]string{sc.name, "nil", fmt.Sprintf("%s | COM:Не удалось записать ключ в БД | Key:%s | ERTX:%v", keyIpUag, sc.subName, err), "1"}
		tp.ExitWithSecTimeout(1)
	}
}

func (ut *SequencesGen) GetCashDetect(keyIpUag string) int {
	data, err := ut.db.Get([]byte(keyIpUag), nil)
	if err == leveldb.ErrNotFound {
		// пусто
		return 0
	} else if err != nil {
		fn := fmt.Sprintf("=>%s", strings.Join([]string{ut.subName, fmt.Sprintf("Get GetCashDetect %s", keyIpUag)}, "=>"))
		ut.loger <- [4]string{ut.name, "GetCashDetect", fmt.Sprintf("%s | Не удалось считать пользователя по keyIpUag из SlowБД | %v", fn, err), "ERROR"}
		tp.ExitWithSecTimeout(1)
	}
	// Данные найдены
	return encod.BytesToInt(data)
}

func (sc *SequencesGen) GetNewUidTable(table []byte) int {
	subName := fmt.Sprintf("=>%s", strings.Join([]string{sc.subName, fmt.Sprintf("Generate new id from %s", table)}, "=>"))
	sicNumBt, err := sc.db.Get(table, nil)
	// Мы исключаем возможности отсутствия таблицы так как при загрузке программы это проверяем
	if err != nil {
		sc.loger <- [4]string{sc.name, "nil", fmt.Sprintf("%s | COM:Не удалось считать из базы | ERTX:%v", subName, err), "1"}
		tp.ExitWithSecTimeout(1)
	}
	sicNum := encod.BytesToInt(sicNumBt)
	fmt.Printf("SequencesGen|NEW|%s|Извлекли|DT:Num:%d\n", table, sicNum)
	sicNum++
	fmt.Printf("SequencesGen|NEW|%s|Инкремент|DT:Num:%d\n", table, sicNum)
	if err := sc.db.Put(table, encod.IntToBytes(sicNum), nil); err != nil {
		sc.loger <- [4]string{sc.name, "nil", fmt.Sprintf("%s | COM:Не удалось записать последовательность %d в таблицу %s | ERTX:%v", subName, sicNum, table, err), "1"}
		tp.ExitWithSecTimeout(1)
	}
	return sicNum
}
