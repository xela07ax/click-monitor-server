package domain


//
//func (g *GenFilter) circle() {
//	rows, err := g.chRepo.Get(&g.filter)
//	if err != nil {
//		g.Loger <- [4]string{"GenFilter.Select", fmt.Sprintf("get[%v|%v]", g.filter.TimestampFrom, g.filter.TimestampTo), fmt.Sprintf("%v", err), "ERROR"}
//		time.Sleep(5 * time.Second)
//		g.circle()
//	}
//	var i int
//	if len(rows) > 0 {
//		g.Loger <- [4]string{"GenFilter.Select", "rows", fmt.Sprintf("extract: %d", len(rows)), "INFO"}
//		for iRow, v := range rows {
//			// проверить есть ли в кеше
//			ipKey := v.Ip.String()
//			rowIpqs := g.db.TableIpAddress.Get(ipKey)
//			if rowIpqs != nil {
//				rowIpqs.RefererId = rowIpqs.Id
//				g.reporting.SetCashDetect(ipKey)
//			} else {
//				i++
//				if i > g.limit {
//					g.Loger <- [4]string{"circle.replaceSender", fmt.Sprintf("CallThirdParty[sender:%s]", g.Sender.HostRepiter), fmt.Sprintf("достигли лимита:%v", g.limit), "INFO"}
//					g.replaceSender()
//				}
//				result, err := g.CallThirdParty(v.Ip.String(), v.UserAgent.String)
//				if err != nil {
//					g.Loger <- [4]string{"circle", "CallThirdParty[ERR_POST]", fmt.Sprintf("ошибка REST [ip:%s][err:%v|resu:%v]", ipKey, err, result), "ERROR"}
//					// если это внутренние ошибки, на будем их регистрировать по правилам конфигурации
//					continue
//				}
//				g.reporting.SenderHost = g.Sender.HostRepiter
//				if isUpdateTariff(result.RespBody) {
//					ertx := fmt.Sprintf("postbackService resp. Error [tip:QUOTA][host:%s][sender:%s][ip:%s]【%s】舞", g.cfg.UrlPostback, g.reporting.SenderHost, ipKey, result.RespBody)
//					if g.cfg.Mock {
//						ertx = fmt.Sprintf("запрос не был отправлен [tip:MOCK][resp: %v]", result)
//					}
//					g.Loger <- [4]string{"circle", "CallThirdParty[ERR_QUOTA]", ertx, "WARNING"}
//					g.reporting.SetErr(ipKey, fmt.Errorf("%s", ertx))
//					g.ErrReq <- struct{}{}
//					continue
//				}
//				g.globalErr = 0
//				g.db.TableIpAddress.SetNew(ipKey, &result)
//				g.reporting.SetOk(ipKey, result.RespBody)
//				g.Loger <- [4]string{"circle", fmt.Sprintf("CallThirdParty[SetOk][%d]", iRow+1), fmt.Sprintf("[ip:%s][uag:%s]", ipKey, v.UserAgent.String)}
//			}
//			time.Sleep(100 * time.Millisecond)
//			continue
//		}
//		//g.db.TableIPQS.SaveBath(ipqsRows)
//	} else {
//		g.Loger <- [4]string{"GenFilter.Select", "rows", "нет строчек для обработки"}
//	}
//
//	<-g.ticker
//	g.calc()
//	g.circle()
//}


//
//func (g *GenFilter) getIndexSender(sender *model.Sender) int {
//	for i, s := range g.cfg.Senders {
//		if s == sender {
//			return i
//		}
//	}
//	panic(fmt.Errorf("не найдена ссылка отправителя"))
//}
//
//func (g *GenFilter) replaceSender() {
//	g.cfg.Mock = true
//	go func() {
//		iSender := g.getIndexSender(g.Sender)
//		iSenderNext := iSender + 1
//		if iSenderNext == len(g.cfg.Senders) {
//			iSenderNext = 0
//		}
//		nextSender := g.cfg.Senders[iSenderNext]
//		g.Loger <- [4]string{"GenFilter.replaceSender", fmt.Sprintf("bad:%s|next:%s", g.Sender.HostRepiter, nextSender.HostRepiter), fmt.Sprintf("sleep:%d minutes START", nextSender.Sleep), "INFO"}
//		time.Sleep(time.Duration(nextSender.Sleep) * time.Minute)
//		g.Loger <- [4]string{"GenFilter.replaceSender", fmt.Sprintf("bad:%s|next:%s", g.Sender.HostRepiter, nextSender.HostRepiter), fmt.Sprintf("sleep:%d minutes END", nextSender.Sleep), "INFO"}
//		if _, ok := g.cashSenders[nextSender]; ok {
//			g.limit = nextSender.ThinkRq
//			delete(g.cashSenders, nextSender)
//		} else {
//			g.limit = nextSender.FirstRq
//			g.cashSenders[nextSender] = struct{}{}
//		}
//		g.Sender = nextSender
//		g.cfg.Mock = false
//	}()
//}
//
//func (g *GenFilter) ErrDaemon() {
//	if len(g.cfg.Senders) < 1 {
//		panic(fmt.Errorf("хостов отправителей не может быть меньше 1-го"))
//	}
//	g.Sender = g.cfg.Senders[0]
//	g.limit = g.Sender.FirstRq
//	g.Loger <- [4]string{"GenFilter.ErrDaemon", "установлен хост отправителя", fmt.Sprintf("sendrHost: %s|limit:%d", g.Sender.HostRepiter, g.limit), "INFO"}
//	g.cashSenders[g.Sender] = struct{}{}
//	go func() {
//		var i int
//		for {
//			<-g.ErrReq // ошибка ответа или другая, нам не ваажно, есть ошибки, переключаем по правилу
//			if g.cfg.Mock {
//				g.Loger <- [4]string{"GenFilter.ErrDaemon", g.Sender.HostRepiter, "IsMock:true (игнорируем ошибку)", "INFO"}
//				continue
//			}
//			if g.globalErr > g.cfg.StopErr {
//				g.Loger <- [4]string{"GenFilter.ErrDaemon", fmt.Sprintf("globalErrPredel:%d", g.globalErr), "достигли максимальное количество ошибок на всех хостах, завершение работы", "WARNING"}
//				tp.ExitWithSecTimeout(1)
//			}
//			i++
//			// смотрим квоту ошибок
//			if i > g.Sender.StopErr {
//				// надо останавливаться или менять отправителя
//				g.globalErr++
//				i = 0
//				g.replaceSender()
//				continue
//			}
//			g.Loger <- [4]string{"GenFilter.ErrDaemon", "resp.error.counter", fmt.Sprintf("зарегистрировано ошибок:%d|предел:%d", i, g.Sender.StopErr), "WARNING"}
//		}
//	}()
//}
//return
//result, err := g.CallThirdParty(v.Ip.String(), v.UserAgent.String)
//if err != nil {
//g.Loger <- [4]string{"circle", "CallThirdParty[ERR_POST]", fmt.Sprintf("ошибка REST [ip:%s][err:%v|resu:%v]", ipKey, err, result), "ERROR"}
//// если это внутренние ошибки, на будем их регистрировать по правилам конфигурации
//continue
//}
//g.reporting.SenderHost = g.Sender.HostRepiter
//if isUpdateTariff(result.RespBody) {
//ertx := fmt.Sprintf("postbackService resp. Error [tip:QUOTA][host:%s][sender:%s][ip:%s]【%s】舞", g.cfg.UrlPostback, g.reporting.SenderHost, ipKey, result.RespBody)
//if g.cfg.Mock {
//ertx = fmt.Sprintf("запрос не был отправлен [tip:MOCK][resp: %v]", result)
//}
//g.Loger <- [4]string{"circle", "CallThirdParty[ERR_QUOTA]", ertx, "WARNING"}
//g.reporting.SetErr(ipKey, fmt.Errorf("%s", ertx))
//g.ErrReq <- struct{}{}
//continue
//}
//g.globalErr = 0
//g.db.TableIpAddress.SetNew(ipKey, &result)
//g.reporting.SetOk(ipKey, result.RespBody)
//g.Loger <- [4]string{"circle", fmt.Sprintf("CallThirdParty[SetOk][%d]", iRow+1), fmt.Sprintf("[ip:%s][uag:%s]", ipKey, v.UserAgent.String)}

//func (g *GenFilter) CallThirdParty(ipAddress, userAgent string) (result model.IPQSRow, err error) {
//	if ipAddress == "" {
//		err = fmt.Errorf("iP address IS EMPTY")
//		return
//	}
//	result.Timestamp = time.Now()
//	result.Uag = userAgent
//	heandler := model.Handle{
//		Time:        result.Timestamp,
//		Send:        !g.cfg.Mock,
//		RedirectUrl: fmt.Sprintf("%s%s/%s", g.cfg.UrlPostback, g.Sender.IpqsKey, ipAddress),
//		Params:      fmt.Sprintf("allow_public_access_points=true&fast=false&lighter_penalties=true&mobile=false&strictness=1&user_agent=%s", strings.ReplaceAll(userAgent, " ", "%20")),
//		Method:      "GET",
//	}
//	dat, err := json.Marshal(heandler)
//
//	respRpc, err := RpcRequest(g.cfg.Service, string(dat), g.Sender.HostRepiter, g.Loger)
//	if respRpc != nil {
//		result.RespStatus = respRpc.RespStatus
//		result.RespCode = respRpc.RespCode
//		result.RespBody = respRpc.RespBody
//	}
//
//	return
//}
//func (g *GenFilter) calc() {
//	g.currentTimestamp = time.Now().Add(- time.Duration(g.cfg.ChDb.Timezone) * time.Hour)
//	g.minusIntervalTimestamp = g.currentTimestamp.Add(- (g.Interval - 1*time.Second))
//	g.filter = model.LogFilter{
//		TimestampFrom: g.minusIntervalTimestamp,
//		TimestampTo:   g.currentTimestamp,
//		Source:        model.Source("redirect"),
//	}
//	g.ticker = time.Tick(g.Interval)
//}
//func isUpdateTariff(text string) bool {
//	indexWarning := strings.Index(text, "\"success\":true") // много ворнингов в сентри по, уберем их оттуда
//	if indexWarning == -1 {
//		return true // если данное вхождение текста (i/o timeout) не найдено, значит это не таймаут а ошибка
//	}
//	return false
//}


//
//func (g *GenFilter) GetFrodIPQSRowsFromFile() {
//	g.Loger <- [4]string{"GenFilter", "GetFrodIPQSRowsFromFile", "froud_ipqs.json"}
//	b, _ := tp.OpenReadFile("froud_ipqs.json")
//	resultupd := make(map[string]model.IPQSRow)
//	upd := make(map[string]model.IPQSRow)
//	json.Unmarshal(b, &resultupd)
//	g.Loger <- [4]string{"LogRepo", "LENTH", fmt.Sprint(len(resultupd))}
//	var bad int
//	var add int
//	for k, v := range resultupd {
//		dbIpqsRow := g.db.TableIpAddress.Get(k)
//		if dbIpqsRow != nil {
//			layout := "2006-01-02T15:04:05.000Z"
//			str := "2022-01-01T11:45:26.371Z"
//			t, _ := time.Parse(layout, str)
//			//fmt.Println(dbIpqsRow)
//			dur := dbIpqsRow.Timestamp.Sub(t)
//			if dur > 10 {
//				bad++
//				fmt.Println(dbIpqsRow.Timestamp.String())
//			}
//		} else {
//			upd[k] = v
//			add++
//		}
//	}
//	fmt.Println("add: ", add)
//	fmt.Println("bad: ", bad)
//	b = nil
//	b, _ = json.Marshal(upd)
//	f, _ := tp.CreateOpenFile("upd.json")
//	f.Write(b)
//	f.Close(); b = nil
//	fmt.Print("End\n")
//	return
//}
//
//func (g *GenFilter) GetFrodRowsFromFile() {
//	g.Loger <- [4]string{"GenFilter", "GetFrodRowsFromFile", "starter"}
//	g.chRepo.GetFrodRowsFromFile()
//}
//func (g *GenFilter) GetFrodLogs() {
//	g.Loger <- [4]string{"GenFilter", "GetFrodRows", "starter"}
//	g.chRepo.GetFrodRows()
//
//}
//
//func (g *GenFilter) ReaDatabase() {
//	dt := g.db.TableIpAddress.ReadAll()
//	str := `api\/json\/ip\/\w+\/(.+)?\?` // нам нужен ip адрес
//	rex := regexp.MustCompile(str)
//	upd := make(map[string]*model.IPQSRow)
//	add := make(map[string]*model.IPQSRow)
//	del := make(map[string]interface{})
//	for _, v := range dt {
//		es := rex.FindStringSubmatch(v.Request)
//		if len(es) < 1 {
//			// если в реквесте вообще нанайдено ip, то удаляем
//			del[v.SenderIp] = struct {}{}
//			continue
//			// если ip в реквести и ip в ключе совпадают
//			//if es[1] == v.SenderIp {
//			//	continue
//			//}
//			// если не совпадают, то тот, что из базы на удаление, так как можем не найти для него Юзерагента
//			// настоящий в список Add
//			//add[es[1]] = v
//			//del = append(del, v.SenderIp)
//		}
//		// если ip в реквести и ip в ключе не совпадают
//		if es[1] != v.SenderIp {
//			//del = append(del, v.SenderIp)
//			del[v.SenderIp] = struct {}{}
//			g.Loger <- [4]string{"ip в ключе не совпадают", es[1], v.SenderIp, "DELETE"}
//			g.Loger <- [4]string{"ip в ключе не совпадают",v.Ip , es[1], "ADD"}
//			v.Ip = es[1]
//			v.SenderIp = ""
//			add[es[1]] = &v
//		}
//		if v.Ip != es[1] {
//			log.Print("IPs")
//			log.Print(v)
//			log.Print(v.Ip)
//			log.Print(v.SenderIp)
//			log.Print(es[1])
//			log.Panic("what")
//		}
//		// проверим UserAgent-ы
//		u, err := url.Parse(v.Request)
//		if err != nil {
//			panic(err)
//		}
//		q, err := url.ParseQuery(u.RawQuery)
//		if err != nil {
//			panic(err)
//		}
//		uagInRequst := q.Get("user_agent")
//		uagInModel := v.Uag
//		if uagInModel != uagInRequst {
//			// если UAG не совпадают, то просто актаализируем
//			g.Loger <- [4]string{"UAG не совпадают", uagInModel,  uagInRequst, "UPDATE"}
//			v.Uag = uagInRequst
//			// сразу добавим на обновление
//			upd[v.Ip] = &v
//		}
//		//
//
//		//v.RpcFields.RespBody = row.ResponseBody
//		//v.Ip = v.SenderIp
//		//v.SenderIp = ""
//		//v.Request = row.RequestUrl
//		//v.Timestamp = row.Timestamp
//		//esResp := rex.FindStringSubmatch(v.Request)
//		//if len(esResp) < 1 {
//		//	panic("that_is_fack")
//		//}
//		//if esResp[1] != v.Ip {
//		//	g.reporting.SetErr(uasg, fmt.Errorf(esResp[1]))
//		//	g.reporting.SetErr(uasg, fmt.Errorf(v.Ip))
//		//	//panic("that_is_fack[ip]")
//		//}
//		//g.reporting.SetOk(uasg, esResp[1])
//		//panic(fmt.Sprintf("not_found ip:%s|%s|%v", address, v.SenderIp, v))
//
//		//fmt.Println(uasg)
//
//		//var req *http.Request
//		//
//		//req, err = http.NewRequest(http.MethodGet, url, nil)
//		//if err != nil {
//		//	err = fmt.Errorf("newRequest[url:%s] error: %v", url, err)
//		//}
//		//query := req.URL.Query()
//
//		//i := 0
//		//for rows.Next() {
//		//	i++
//		//	r.Loger <- [4]string{"LogRepo.scanRowsFrodTable", fmt.Sprintf("rows.Next[%d]", i), "scan start"}
//		//	log := FrodRow{}
//		//	err := rows.Scan(
//		//		&log.Timestamp,
//		//		&log.RequestUrl,
//		//		&log.ResponseBody,
//		//		&log.Ip,
//		//		&log.UserAgent,
//		//	)
//		//	if err != nil {
//		//		panic(err)
//		//	}
//		//
//		//	es := rex.FindStringSubmatch(log.RequestUrl)
//		//	if len(es) > 1 {
//		//		if _, ok := cashIpqsKey[es[1]]; !ok {
//		//			cashIpqsKey[es[1]] = true
//		//			r.SetLogIpqs(es[1])
//		//		}
//		//	} else {
//		//		r.SetLogIpqs(fmt.Sprintf("not_found:%s", log.RequestUrl))
//		//	}
//		//	r.WorkerScanner(log)
//		//	r.Loger <- [4]string{"LogRepo.scanRowsFrodTable", fmt.Sprintf("rows.Next[%d]", i), "scan WorkerScanner_ok"}
//		//
//		//}
//		//r.Loger <- [4]string{"LogRepo.scanRowsFrodTable", "END__", "scan finish"}
//		//fmt.Println("End scan")
//	}
//
//	var b []byte
//	var f *os.File
//	g.Loger <- [4]string{"upd", "LENGTH", fmt.Sprint(len(upd))}
//	// upd := make(map[string]*model.IPQSRow)
//	b, _ = json.Marshal(upd)
//	f, _ = tp.CreateOpenFile("upd.json")
//	f.Write(b)
//	f.Close(); b = nil
//	g.Loger <- [4]string{"add", "LENGTH", fmt.Sprint(len(add))}
//	// add := make(map[string]*model.IPQSRow)
//	b, _ = json.Marshal(add)
//	f, _ = tp.CreateOpenFile("add.json")
//	f.Write(b)
//	f.Close(); b = nil
//	g.Loger <- [4]string{"del", "LENGTH", fmt.Sprint(len(del))}
//	// del := make([]string, 0, 16)
//	b, _ = json.Marshal(del)
//	f, _ = tp.CreateOpenFile("del.json")
//	f.Write(b)
//	f.Close(); b = nil
//	fmt.Println(rex, " Gooodby!")
//}


//
//func (g *GenFilter) circle() {
//	rows, err := g.chRepo.Get(&g.filter)
//	if err != nil {
//		g.Loger <- [4]string{"GenFilter.Select", fmt.Sprintf("get[%v|%v]", g.filter.TimestampFrom, g.filter.TimestampTo), fmt.Sprintf("%v", err), "ERROR"}
//		time.Sleep(5 * time.Second)
//		g.circle()
//	}
//	var i int
//	if len(rows) > 0 {
//		g.Loger <- [4]string{"GenFilter.Select", "rows", fmt.Sprintf("extract: %d", len(rows)), "INFO"}
//		for iRow, v := range rows {
//			// проверить есть ли в кеше
//			ipKey := v.Ip.String()
//			rowIpqs := g.db.TableIpAddress.Get(ipKey)
//			if rowIpqs != nil {
//				rowIpqs.RefererId = rowIpqs.Id
//				g.reporting.SetCashDetect(ipKey)
//			} else {
//				i++
//				if i > g.limit {
//					g.Loger <- [4]string{"circle.replaceSender", fmt.Sprintf("CallThirdParty[sender:%s]", g.Sender.HostRepiter), fmt.Sprintf("достигли лимита:%v", g.limit), "INFO"}
//					g.replaceSender()
//				}
//				result, err := g.CallThirdParty(v.Ip.String(), v.UserAgent.String)
//				if err != nil {
//					g.Loger <- [4]string{"circle", "CallThirdParty[ERR_POST]", fmt.Sprintf("ошибка REST [ip:%s][err:%v|resu:%v]", ipKey, err, result), "ERROR"}
//					// если это внутренние ошибки, на будем их регистрировать по правилам конфигурации
//					continue
//				}
//				g.reporting.SenderHost = g.Sender.HostRepiter
//				if isUpdateTariff(result.RespBody) {
//					ertx := fmt.Sprintf("postbackService resp. Error [tip:QUOTA][host:%s][sender:%s][ip:%s]【%s】舞", g.cfg.UrlPostback, g.reporting.SenderHost, ipKey, result.RespBody)
//					if g.cfg.Mock {
//						ertx = fmt.Sprintf("запрос не был отправлен [tip:MOCK][resp: %v]", result)
//					}
//					g.Loger <- [4]string{"circle", "CallThirdParty[ERR_QUOTA]", ertx, "WARNING"}
//					g.reporting.SetErr(ipKey, fmt.Errorf("%s", ertx))
//					g.ErrReq <- struct{}{}
//					continue
//				}
//				g.globalErr = 0
//				g.db.TableIpAddress.SetNew(ipKey, &result)
//				g.reporting.SetOk(ipKey, result.RespBody)
//				g.Loger <- [4]string{"circle", fmt.Sprintf("CallThirdParty[SetOk][%d]", iRow+1), fmt.Sprintf("[ip:%s][uag:%s]", ipKey, v.UserAgent.String)}
//			}
//			time.Sleep(100 * time.Millisecond)
//			continue
//		}
//		//g.db.TableIPQS.SaveBath(ipqsRows)
//	} else {
//		g.Loger <- [4]string{"GenFilter.Select", "rows", "нет строчек для обработки"}
//	}
//
//	<-g.ticker
//	g.calc()
//	g.circle()
//}
//func cleanHost(host string) string {
//	host = strings.ReplaceAll(host,"http://","")
//	host = strings.ReplaceAll(host,".","")
//	host = strings.ReplaceAll(host,":","")
//	host = strings.ReplaceAll(host,"/rpc","")
//	return host
//}
// deprecate!
//gen.GetFrodIPQSRowsFromFile()
//gen.GetFrodRowsFromFile()
//tp.ExitWithSecTimeout(0)
//return gen
//if cfg.ModeStart.ClickMonitor {
//	gen.ErrDaemon()
//	gen.calc()
//	go gen.circle()
//	return gen
//}