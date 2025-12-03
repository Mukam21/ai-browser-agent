package browser

import (
	"context"
	"log"
	"runtime"
)

type Controller interface {
	Navigate(url string) error
	Click(selector string) error
	Fill(selector, value string) error
	Screenshot(path string) error
	GetPageState() (string, error)
	GetDOM() (string, error)
	Close()
}

func NewController(ctx context.Context) (Controller, error) {
	log.Println(" Инициализирую контроллер браузера...")

	if runtime.GOOS == "windows" {
		log.Println("🛡️ Windows обнаружена, использую безопасный контроллер")
		return NewSafeController(ctx), nil
	}

	log.Println(" Не-Windows ОС, пробую Rod контроллер...")
	return tryRodController(ctx)
}

func tryRodController(ctx context.Context) (Controller, error) {
	safeCtrl := NewSafeController(ctx)

	log.Println("Rod контроллер временно отключен, использую безопасный")
	return safeCtrl, nil
}
