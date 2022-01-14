package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/xela07ax/rest-repiter/model"
)

//
//import (
//	"bytes"
//	"fmt"
//	"github.com/niubaoshu/gotiny"
//	"github.com/syndtr/goleveldb/leveldb"
//	"github.com/xela07ax/rest-repiter/model"
//	"github.com/xela07ax/toolsXela/encod"
//	"strings"
//	"time"
//)
//
type UserAgent struct {
	name string
	subName string
	sequences SequencesGen
	config *model.Config
	dbLoginUid *leveldb.DB
	db *leveldb.DB
	loger chan<- [4]string
}
//func (ut *UserAgent) ReadAll() {
//	fmt.Printf("ReadAll|Start\n")
//	fmt.Printf("ReadAll|MTable|Iterator\n")
//	fn := fmt.Sprintf("=>%s",strings.Join([]string{ut.subName,"Read all uaseragent"},"=>"))
//	iter := ut.dbLoginUid.NewIterator(nil, nil)
//	for iter.Next() {
//		fmt.Printf("ReadAll|MTable|DT:Key:%s|DT:Val:%d\n",iter.Key(),encod.BytesToInt(iter.Value()))
//	}
//	iter.Release()
//	err := iter.Error()
//	if err != nil {
//		ut.loger<- [4]string{ut.name,"nil",fmt.Sprintf("%s | COM:Не удалось считать в UserAgent Iterate-ом | ERTX:%v",fn,err),"1"}
//		ut.config.ExitProgramErr()
//		return
//	}
//	fmt.Printf("ReadAll|Table|Iterator\n")
//	iter = ut.db.NewIterator(nil, nil)
//	for iter.Next() {
//		var user model.UserDetail
//		gotiny.Unmarshal(iter.Value(), &user)
//		fmt.Printf("ReadAll|Table|DT:Key:%v|DT:Val:%v\n",encod.BytesToInt(iter.Key()),user)
//	}
//	iter.Release()
//	err = iter.Error()
//	if err != nil {
//		ut.loger<- [4]string{ut.name,"nil",fmt.Sprintf("%s | COM:Не удалось считать в UserAgent Iterate-ом | ERTX:%v",fn,err),"1"}
//		ut.config.ExitProgramErr()
//		return
//	}
//	return
//}
//func (ut *UserAgent) SaveBathUids(uids map[string]int) {
//	fmt.Printf("SaveBatch|%v\n",uids)
//	if len(uids) == 0 {
//		return
//	}
//	fn := fmt.Sprintf("=>%s",strings.Join([]string{ut.subName,"Save bath UIDs"},"=>"))
//	batch := new(leveldb.Batch)
//	for login, uid := range uids {
//		fmt.Printf("SaveBatch|UID:%d|login:%s\n",uid,login)
//		batch.Put([]byte(login), encod.IntToBytes(uid))
//	}
//	err := ut.dbLoginUid.Write(batch,nil)
//	if err != nil {
//		ut.loger<- [4]string{ut.name,"nil",fmt.Sprintf("%s | COM:Не удалось записать в IndexUidLogin Batch-ем | ERTX:%v",fn,err),"1"}
//		ut.config.ExitProgramErr()
//	}
//}
//func (ut *UserAgent) SaveBathUsers(uaseragent []model.UserDetail) {
//	fn := fmt.Sprintf("=>%s",strings.Join([]string{ut.subName,"Save bath Users"},"=>"))
//	batch := new(leveldb.Batch)
//	for _, user := range uaseragent {
//		batch.Put(encod.IntToBytes(user.UID), gotiny.Marshal(&user))
//	}
//	err := ut.db.Write(batch,nil)
//	if err != nil {
//		ut.loger<- [4]string{ut.name,"nil",fmt.Sprintf("%s | COM:Не удалось записать в UserAgent Batch-ем | ERTX:%v",fn,err),"1"}
//		ut.config.ExitProgramErr()
//	}
//}
//
//func (ut *UserAgent) ConvertUidToLogin(uid []byte)(login []byte) { // Можно проверять через nil
//	fmt.Printf("=>IterateConvertUidLogin| Попытка найти UID|DT:UID:%d\n",encod.BytesToInt(uid))
//
//	iter := ut.dbLoginUid.NewIterator(nil, nil)
//	for iter.Next() {
//		// Найдем логин
//		// Comparing slice
//		// Using Compare function
//		fmt.Printf("=>IterateConvertUidLogin| DT:val:%d\n",iter.Value())
//		res := bytes.Compare(iter.Value(), uid)
//		if res == 0 { //Срез равен
//			login = iter.Value()
//			break
//		}
//	}
//	iter.Release()
//	time.Sleep(1*time.Second)
//	return
//}
//func (ut *UserAgent) GetUserLogin(login string)model.UserDetail {// если вернется .UID == 0 значит не найдено
//	// Нужно узнать его Uid
//	uidBt, err := ut.dbLoginUid.Get([]byte(login), nil)
//	fmt.Printf("=>GetUserLogin| DT:login:%s|DT:UIDBT:%v\n",login,uidBt)
//	if err == leveldb.ErrNotFound {
//		return model.UserDetail{}
//	}
//	if err != nil {
//		fn := fmt.Sprintf("=>%s",strings.Join([]string{ut.subName,fmt.Sprintf("Get UID %s",login)},"=>"))
//		ut.loger<- [4]string{ut.name,login,fmt.Sprintf("%s | Не удалось считать Uid из БД | %v",fn,err),"1"}
//		ut.config.ExitProgramErr()
//	}
//	// Uid получен, запрашиваем карточку
//	return ut.getUserUid(uidBt)
//}
//
//func (ut *UserAgent) getUserUid(uid []byte)model.UserDetail {
//	userBt, err := ut.db.Get(uid, nil)
//	if err != nil {
//		uid := encod.BytesToInt(uid)
//		fn := fmt.Sprintf("=>%s",strings.Join([]string{ut.subName,fmt.Sprintf("Get UserDetail %d",uid)},"=>"))
//		ut.loger<- [4]string{ut.name,"nil",fmt.Sprintf("%s | Не удалось считать пользователя по Uid из БД (возможно его Uid в таблице нет, но Login в базе есть)* | %v",fn,err),"1"}
//		ut.config.ExitProgramErr()
//	}
//	// Данные найдены
//	var user model.UserDetail
//	gotiny.Unmarshal(userBt, &user)
//	return user
//}
