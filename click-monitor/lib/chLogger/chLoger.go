package chLogger

import (
	"fmt"
	"github.com/xela07ax/toolsXela/tp"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	ConsolFilterFn map[string]int // map[funcName]mode
	ConsolFilterUn map[string]int // map[unitName]mode
	Mode           int
	Dir            string // Папка для сохранений
	Broadcast      chan<- []byte
}


type mutexStopper struct {
	mu sync.Mutex
	x  bool
}

func (c *mutexStopper) stopSignal() {
	c.mu.Lock()
	c.x = true
	c.mu.Unlock()
}

func (p *ChLoger) Stop() {
	close(p.done)
	for i := 0; i < p.GetCores(); i++ {
		<-p.stopX
	}
	fmt.Println(tp.Getime(), "| ** END ** | Chan Logger |")
	return
}

func (p *ChLoger) SignalStoper(prepareExit <-chan bool) {
	<-prepareExit
	p.Stop()
	p.GoodBy <- true
	fmt.Println(tp.Getime(), "| ** GoodBy <- END ** | Chan Logger |")
	return
}

type ChLoger struct {
	Options  *Config
	batchCnt *mutexCounter
	wg  sync.WaitGroup
	process *mutexRunner
	ChInLog  chan [4]string //Приемник строк
	done   chan struct{} //Сигнал, что пока закругляться
	stopX  chan bool     //Миньон завершила свою работу
	GoodBy chan bool
}

func NewChLoger(cfg *Config) *ChLoger {
	time := tp.Getime()
	time = strings.Replace(time, " ", "_", -1)
	time = strings.Replace(time, ":", "", -1)
	tp.FckText(fmt.Sprintf("Создание папки логов: %s", cfg.Dir), tp.CheckMkdir(cfg.Dir))
	return &ChLoger{Options: cfg, ChInLog: make(chan [4]string, 100), done: make(chan struct{}), wg: sync.WaitGroup{}, stopX: make(chan bool, 1000), batchCnt: new(mutexCounter), process: new(mutexRunner), GoodBy: make(chan bool)}
}

type mutexRunner struct {
	mu sync.Mutex
	x  int
}
func (c *mutexRunner) addRun()(i int) {
	c.mu.Lock()
	c.x++
	i = c.x
	c.mu.Unlock()
	return
}
func (c *mutexRunner) getCores()(cores int)   {
	c.mu.Lock()
	cores = c.x
	c.mu.Unlock()
	return
}

func (c *ChLoger) GetCores()(cores int)  {
	cores = c.process.getCores()
	return
}


func (p *ChLoger) RunMinion() {
	go p.runMinion(p.process.addRun())
	return
}

type mutexCounter struct {
	mu sync.Mutex
	x  int
}

func (c *mutexCounter) cntPlus() {
	c.mu.Lock()
	c.x++
	c.mu.Unlock()
}

func (c *ChLoger) GetCountPackage() (cnt int) {
	c.batchCnt.mu.Lock()
	cnt = c.batchCnt.x
	c.batchCnt.mu.Unlock()
	return
}

//ch <- chan [3][]string //Приемник строк
//[0] - Имя функции
//[1] - Имя Etl
//[2] - Текст
//[3] - Тип "error","ok"

func (p *ChLoger) exec(gopher int, elem [4]string) {
	//var etlFlush string
	var funcFlush string
	ln := len(elem[2])-1
	if elem[2][ln:][0] == 10 {
		elem[2] =elem[2][:ln]
	}

	//fmt.Println(elem[len(elem[2])-2:])
	// Посмотрим есть ли эта функция в правилах
	if val, ok := p.Options.ConsolFilterFn[elem[0]]; ok {
		// Если режим совпадает, то печатаем  или скрываем
		if p.Options.Mode != val {
			return
		}
		// Посмотрим есть ли этот юнит в правилах
		if val, ok := p.Options.ConsolFilterUn[elem[1]]; ok {
			// Если режим совпадает, то печатаем  или скрываем
			if p.Options.Mode != val {
				return
			}
		}
	}
	//etlFlush = fmt.Sprintf("%s | FUNC: %s | TEXT: %s\n",tp.Getime(),elem[0],elem[2])
	funcFlush =fmt.Sprintf("%s|FUNC:%s|UNIT:%s|%s|TEXT: %s\n",tp.Getime(),elem[0],elem[1],elem[3],elem[2])
	if elem[1] != "nil" { // Если юнит не задан
		//fileEtl := fmt.Sprintf("%s.txt",elem[1])
		//fEtl, err := tp.OpenWriteFile(filepath.Join(p.Options.Dir,fileEtl))
		//tp.FckText(fmt.Sprintf("Логгер | Открытие файла: %s",fileEtl),err)
		//fEtl.Write([]byte(etlFlush))
		//fEtl.Close()
	}
	if elem[3] == "1" || elem[3] == "ERROR" { // Если ошибка
		errFlush := fmt.Sprintf("%s | UNIT: %s | FUNC: %s | TEXT: %s\n",tp.Getime(),elem[1],elem[0],elem[2])
		fileEtl := fmt.Sprintf("%s.txt","Errors")
		fEtl, err := tp.OpenWriteFile(filepath.Join(p.Options.Dir,fileEtl))
		tp.FckText(fmt.Sprintf("Логгер | Открытие файла: %s",fileEtl),err)
		fmt.Fprintf(os.Stderr, errFlush)
		if p.Options.Broadcast != nil {
			p.Options.Broadcast <- []byte(fmt.Sprintf("***ERRX*** %s",errFlush))
		}
		fEtl.Write([]byte(errFlush))
		fEtl.Close()
	} else {
		fmt.Print(funcFlush)
		if p.Options.Broadcast != nil {
			p.Options.Broadcast <- []byte(fmt.Sprintf("***OKAY*** %s",funcFlush))
		}
	}
	fileFunc := fmt.Sprintf("%s.txt",elem[0])
	fFunc, err := tp.OpenWriteFile(filepath.Join(p.Options.Dir,fileFunc))
	tp.FckText(fmt.Sprintf("Логгер | Открытие файла: %s",fileFunc),err)
	fFunc.Write([]byte(funcFlush))
	fFunc.Close()
}
func (p *ChLoger) runMinion(gopher int) {
	doneRutine := make(chan struct{})
	p.circle(gopher, doneRutine)
}

func (p *ChLoger) circle(gopher int, doneRutine chan struct{}) {
	for {
		select {
		case element := <-p.ChInLog:
			p.exec(gopher, element)
		case <-p.done:
			if len(p.ChInLog) == 0 {
				p.stopX <- true
				return
			}
		}
	}
}
