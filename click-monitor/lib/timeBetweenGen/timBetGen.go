package timeBetweenGen

import (
	"fmt"
	"time"
)

type BetweenTimer struct {
	currentTimestamp       time.Time
	timeZoneHour           int
	minusIntervalTimestamp time.Time
	intervalSecond         int
	Ticket                 chan [2]time.Time
	Loger                  chan<- [4]string
}

func calc(timeZoneHour, intervalSec int) (currentTimestamp, minusIntervalTimestamp time.Time) {
	currentTimestamp = time.Now().Add(- time.Duration(timeZoneHour) * time.Hour)
	return currentTimestamp, currentTimestamp.Add(- (time.Duration(intervalSec-1) * time.Second))
}

func NewBetweenTimer(timeZoneHour, intervalSec int, loger chan<- [4]string) *BetweenTimer {
	bTimer := &BetweenTimer{
		timeZoneHour:   timeZoneHour,
		intervalSecond: intervalSec,
		Ticket: make(chan [2]time.Time),
		Loger: loger,
	}
	return bTimer
}

func (b *BetweenTimer) GetBetweenDates() (currentTimestamp, minusIntervalTimestamp time.Time) {
	return calc(b.timeZoneHour, b.intervalSecond)
}

func (b *BetweenTimer) StartDaemonTicker() {
	ticker := time.Tick(time.Duration(b.intervalSecond) * time.Second)
	// что-бы сразу запустился процесс сначала сгенерируем даты и отправим их в канал
	stPrep, endPrep := calc(b.timeZoneHour, b.intervalSecond)
	b.Ticket <- [2]time.Time{stPrep, endPrep}
	b.Loger <- [4]string{"BetweenTimer.StartDaemonTicker", "таймер генератор sql between - запущен", fmt.Sprintf("intervalSecond:%d", b.intervalSecond)}
	go func(out chan<- [2]time.Time, tick <-chan time.Time) {
		for {
			<-tick
			st, end := calc(b.timeZoneHour, b.intervalSecond)
			select {
			case out <- [2]time.Time{st, end}:
			case <-tick:
				b.Loger <- [4]string{"BetweenTimer", "DaemonTicker", "никто не заберал _тик_ раньше следующего", "WARNING"}
				// если никто не заберит "тик", то следующий "тик" будет разблокирующим
			}
		}
	}(b.Ticket, ticker)
}
