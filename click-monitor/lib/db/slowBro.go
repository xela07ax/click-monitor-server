package db

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/xela07ax/rest-repiter/model"
	"github.com/xela07ax/toolsXela/tp"
	"path/filepath"
	"sync"
	"time"
)

type (
	dictTables struct {
		sequence              string
		ipaddress             string
		convertDialogIDToUids string
		uaseragent            string
		ipqs                  string
		fileStore             string
	}
	Slowpoke struct {
		slowpokeName   string //Какие то данные которые нужны для бизнеса
		config         *model.Config
		tables         map[string]*leveldb.DB
		tablesName     dictTables
		Sequences      SequencesGen
		TableUserAgent UserAgent
		TableIpAddress IpAddressTable
		TableIPQS      IPQSTable
		delayMs        time.Duration
		loger          chan<- [4]string
		process        *mutexRunner  // Системная функция конвеера, для открытия нового миньона без гонок
		ChSlowIN       chan int      // Приемник данных, с которыми сервис будет работать
		stopPrepare    *mutexStopper // Сигнал всем миньонам, что пока закругляться
		stopX          chan bool     // Миньон отправляет сигнал сюда, о завершении своей работы и уничтожения. Когда все отправят сигнал коннвеер прекратит работу
		GoodBy         chan bool     // Внешний сигнал, о том, что конвеер прекратил работу
	}
)

func (app *Slowpoke) CloseStore() {
	var err error
	for name, path := range app.tables {
		app.loger <- [4]string{app.slowpokeName, "nil", fmt.Sprintf("COM:Закрываем базу %s\n", name)}
		err = path.Close()
		if err != nil {
			app.loger <- [4]string{app.slowpokeName, "nil", fmt.Sprintf("=>Close store | COM:Закрытие в BD %s не удалось: %v\n", name, err), "1"}
		}
	}
}

//dbSessions, dbDialos, dbSequence,logic, loger,200
func NewStore(config *model.Config, loger chan<- [4]string) *Slowpoke { // <- Вход в библиотеку тут!
	name := "Slowpoke DB"
	// Я так понимаю. этот модуль будет потихоньку записывать данные в базу
	loger <- [4]string{name, "nil", "Добро пожаловать в управление локальным хранилищем, буду записывать данные в базу"}
	// Открываем базу данных
	//tp.CheckMkdir(filepath.Join(config.Path.BinDir, "levelDB"))
	var errGen = func(name, path string, err error) {
		if err != nil {
			loger <- [4]string{"Open LevelDB storage", "nil", fmt.Sprintf("Не удалось открыть хранилище %s по пути %s | %v\n", name, path, err), "1"}
			tp.ExitWithSecTimeout(1)
		}
	}
	var pathGen = func(path string) string {
		return filepath.Join(config.Path.BinDir, path)
	}
	var err error
	tables := make(map[string]*leveldb.DB)
	dt := dictTables{
		sequence:   "FileSequence",
		ipaddress:  config.Reporting.TableIPaddress,
		uaseragent: config.Reporting.TableUserAgent,
		ipqs:       config.Reporting.DbPath_Ipqs,
	}
	for nameTable, path := range map[string]string{
		dt.sequence:   pathGen(config.Reporting.DbPath_sequence),
		dt.ipaddress:  pathGen(config.Reporting.DbPath_IPaddress),
		dt.uaseragent: pathGen(config.Reporting.DbPathUserAgent),
		dt.ipqs:       pathGen(config.Reporting.DbPath_Ipqs),
	} {
		loger <- [4]string{name, "nil", fmt.Sprintf("Открываем таблицу %s | DT:%s", nameTable, path)}
		tables[nameTable], err = leveldb.OpenFile(path, nil)
		errGen(nameTable, path, err)
	}
	// Последовательность может быть пустая, нельзя, этого допустить
	sc := SequencesGen{
		name:    name,
		subName: "Sequences Generator",
		Tables: NameTables{
			IPQS: []byte(config.Reporting.DbPath_Ipqs),
			UAG:  []byte(config.Reporting.DbPathUserAgent),
			IP:   []byte(config.Reporting.DbPath_IPaddress),
		},
		config: config,
		db:     tables[dt.sequence],
		loger:  loger,
	}
	sc.sequenceInit()
	ut := UserAgent{
		name:       name,
		subName:    "UserAgent Table",
		sequences:  sc,
		config:     config,
		dbLoginUid: tables[dt.ipaddress],
		db:         tables[dt.uaseragent],
		loger:      loger,
	}
	dtb := IpAddressTable{
		name:             name,
		subName:          "IpAddress Table",
		sequences:        sc,
		config:           config,
		dbDIDtoClientUid: tables[dt.convertDialogIDToUids],
		db:               tables[dt.ipaddress],
		loger:            loger,
	}
	ipqs := IPQSTable{
		name:             name,
		subName:          "IPQSTable Table",
		sequences:        sc,
		config:           config,
		dbDIDtoClientUid: tables[dt.convertDialogIDToUids],
		db:               tables[dt.ipqs],
		loger:            loger,
	}

	return &Slowpoke{
		slowpokeName:   name, //Присваеваем входные данные для использования всем миньонам
		config:         config,
		tables:         tables,
		tablesName:     dt,
		Sequences:      sc,
		TableUserAgent: ut,
		TableIpAddress: dtb,
		TableIPQS:      ipqs,
		delayMs:        time.Duration(config.Reporting.DbSlowWriterDaley_Ms) * time.Millisecond, // Задержка считывания очереди записи в базу, ссылка для того, что бы динамично управлять скоростью
		loger:          loger,                                                                   //Присваеваем входные данные для использования всем миньонам
		ChSlowIN:       make(chan int, 1000),                                                    // Входной канал можно передавать из вне, но если начало должно происходить внутри этого конвеера, то можно ее убрать вообще из использования и иницализировать только вывод
		stopPrepare:    new(mutexStopper),                                                       // Системная переменная, инициализация внутри
		process:        new(mutexRunner),                                                        // Системная переменная, инициализация внутри
		stopX:          make(chan bool, 1000),                                                   // Системная переменная, инициализация внутри
		GoodBy:         make(chan bool, 2),                                                      // Системная переменная, инициализация внутри
	} // Отправляем ссылку на объект для его управления и слежения
}

// Область функций корректного завершения
// Более полное описание https://habr.com/ru/post/271789/
type mutexStopper struct {
	mu sync.Mutex // <-- этот мьютекс защищает от гонок
	x  bool       // Cтатус, пора ли нам выходить когда входной канал опустел
}

// Внутренняя безопасная функция для установки статса "Пора закругляться" всем миньонам
func (c *mutexStopper) stopSignal() {
	c.mu.Lock()         // Блокируем все операции, с этой функцией, остальные ждут завершения
	defer c.mu.Unlock() // Отложеная разблокировка
	c.x = true          // Устанавливаем заветный статус
	return              // <-- разблокировка произойдет здесь, defer выполняются при выходе из функции
}

// Две публичные функции для завершения
// Первая
// вызывается в синхронном режиме (без приставки GO)
// Посылается сигнал "Стоп" всем миньонам, когда входящий канал опустеет и они все корректно пмрут, эта фкнкция перестанет держать вызывающую область.
// В Асинхронном режиме позволяет в фоновом подать сигнал закругляться, и сообщит о своей остановке в GoodBy
// В синхронном держит тело вызывающей фкнкции до корректного завершения, так же посылает сигнал в GoodBy
func (p *Slowpoke) Stop() {
	p.stopPrepare.stopSignal()          // Устанавливаем статус завершения, что бы все миньены знали о готовящемся выходе
	for i := 0; i < p.GetCores(); i++ { // Джем подтверждения корректного завершения всех запущеных миньонов, количество запущеных отражается в p.GetCores()
		<-p.stopX // Получаем один сигнал, от одной рутины, значит она закончила работу
	}
	p.GoodBy <- true //Посылаем сигнал, что конвеер завершил все в канал. Такой же конывеер может получить этот сигнал и будет собирать манатки
	fmt.Println("Модуль-балансировщик завершил все миньоны")
	return
}

// Вторая
// вызывается как с приставкой GO, так и без
// На вход ждет канал с сигналом "Пора закругляться"
// Когда сигнал поступает, и мрут все миньоны посылает сигнал в публичный канал GoodBy.
// В Асинхронном режиме позволяет в фоновом режиме узнавать, что надо закругляться и сообщит о своей остановке
// В синхронном держит программу до получения сигнала и корректного завершения, так же посылает сигнал в GoodBy
func (p *Slowpoke) SignalStoper(off <-chan bool) {
	<-off                                                    // Ждем сигнал true, для начала завершения программы
	p.Stop()                                                 // Сигнал получен, процесс пошел, прогресс не остановить. Будем тут пока все не кончится
	fmt.Println("Модуль-балансировщик завершил все миньоны") //Просто печатаем, что б видеть ход работы
	p.GoodBy <- true                                         //Посылаем сигнал, что конвеер завершил все в канал. Такой же конывеер может получить этот сигнал и будет собирать манатки
	return                                                   // Данный конвеер завершил свою работу. Остаются жить только присвоеные в объект данные за время жизни, пока они кому нибудь нужны из вне.
}

// Область функций корректного завершения здесь заканчивается

// Область для безопасного открытия миньюна
// Более полное описание https://habr.com/ru/post/271789/
type mutexRunner struct {
	mu sync.Mutex // <-- этот мьютекс защищает инкремент ниже
	x  int        // <-- это поле под ним
}

func (c *mutexRunner) addRun() (i int) {
	c.mu.Lock() // Блокируем выполнение все операции, что бы код в нутри выполнялся только этот, а остальные ждали завершения
	defer c.mu.Unlock()
	c.x++   //Простая операция инкремента
	i = c.x // Обновляем информацию об объекте
	// какой-нибудь интересный код, который можно выполнять без страха гонок
	// <-- defer не будет выполнен тут, как кто-нибудь *может* подумать
	return // <-- разблокировка произойдет здесь, defer выполняются при выходе из функции
}

//Мункция, для запроса охраняемого значения, например тут количество запущеных миньончиков
func (c *Slowpoke) GetCores() (cores int) { // Сразу объявляем еременную на вывод
	c.process.mu.Lock()   // Блокируем выполнение все операции, что бы ничего не могло помешать
	cores = c.process.x   // Считываем количество запущеным миньонов и присваеваем в выходную переменную
	c.process.mu.Unlock() // <-- разблокировка произойдет здесь
	return                // Поскольку выходная переменная уже объявлена и заполнена, можно ничего тут больше не писать. Го все понял и функция завершает работу.
}

// Собственно сам внешний вызов, для безопасного открытия миньюна
func (p *Slowpoke) RunMinion() {
	go p.minion(p.process.addRun())
	// Так же нам возвращается порядковый номер миньона, которы номеруется с 1-го и передается в конвеер.
	// Rак видно, ничего держать эту функцию не будет, она асинхронна
	// Миньон может быть еще не создан, но работа над этим идет
	return
}

// Область для безопасного открытия миньюна заканчивается

// Все самое интересное начинается тут
func (p *Slowpoke) minion(gophere int) {
	// fmt.Printf("Миньен %d инициализирован\n",gophere) //Печатаем ход работ.
	/*
		Какой то код, который нужно выполнить до того как начнется сам процесс считывания данных со входящкго канала
	*/
	for { // Начинается тело сканирования каналов, из цикла выходить не планируем, завершение цикла будет завершение функции
		select { // Смотрим в какие каналы пришли данные и считываем с него
		case elem := <-p.ChSlowIN: // Канал данных, с которыми нужно проводить какие то операции, к примеру строки или уже обработаные другим конвеером строки
			fmt.Println(elem)
			p.loger <- [4]string{p.slowpokeName, "nil", fmt.Sprintf("Очереди ChSlowIN: %d |  Предел: 1000 | Задержка Мс: %d | Работников: %d", len(p.ChSlowIN), p.config.Reporting.DbSlowWriterDaley_Ms, p.process.x)}
		default: // Попадаем сюда когда нет работы, канал пуст и тут можем узнать, может пора закругляться
			if p.stopPrepare.x { // Функции сверху обрабатывают два типа сигнала, это по прихода из канала или просто вызов Stop, и они передают сюда true, значит нам пора выключаться
				// Какой то интересный завершающий код
				fmt.Printf("Завершаю работу, я это миньен %d, имя: %s\n", gophere, p.slowpokeName) //Печатаем ход работ. И в тексте добавляем полезную информацию, о количестве завершеных им операций и порядковый номер миньона
				p.stopX <- true                                                                    // Посылаем сигнал, что миньон прибит
				return                                                                             // Собственно ей конец
			}
			// Это сделано, что бы лишний раз не дергать такт процессора на проверку, почему входной канал пуст, если пуст и умирать рано, подождем секундочку
		}
		time.Sleep(p.delayMs * time.Millisecond) // Будем писать в базу потихоньку!!! Осторожно, нужно следить за очередью, может переполниться
	}
}
