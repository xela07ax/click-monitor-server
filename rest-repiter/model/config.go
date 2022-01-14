package model

import (
	"fmt"
	"os"
	"syscall"
)

type Pathes struct {
	Log    string `json:"log"`
	Report string `json:"report"`
	BinDir string `json:"_,omitempty"`
}

type Config struct {
	HttpServerPort int    // Порт программы
	Path           Pathes `json:"path"`
	Service        string `json:"service"`
	Timeout        int    `json:"timeout"`
	// В конфиге не проставлять!
	KeySession []byte         `json:"omitempty"`
	ChExit     chan os.Signal `json:"_,omitempty"`
}

func (cfg *Config) ExitProgramErr() {
	cfg.exitProgram(1)
}
func (cfg *Config) ExitProgramNorm() {
	cfg.exitProgram(0)
}
func (cfg *Config) exitProgram(status int) {
	// Го любит завершать свою работу раньше чем сделать все завершающие операции, но все же он остается очень быстрым если убрать задержку и немного подождать
	// 0 - norm
	// 1 - error
	fmt.Println("Завершение работы программы, ускоряем выдачу логов")
	// time.Sleep(1*time.Second)
	// Внимание!, обработка статусов выхода временно недоступно, выход всегда безошибочный*
	//os.Exit(status)
	// Отправляем сигнал завершения в функцию безопасного выхода "Daemon"
	cfg.ChExit <- syscall.SIGSTOP
}
