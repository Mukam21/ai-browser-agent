package browser

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type SafeController struct {
	ctx        context.Context
	currentURL string
}

func NewSafeController(ctx context.Context) *SafeController {
	return &SafeController{
		ctx: ctx,
	}
}

func (sc *SafeController) Navigate(url string) error {
	log.Printf(" Открываю: %s", url)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("неподдерживаемая ОС: %s", runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("не удалось открыть браузер: %w", err)
	}

	sc.currentURL = url
	time.Sleep(3 * time.Second)

	log.Printf("✅ URL открыт в системном браузере: %s", url)
	return nil
}

func (sc *SafeController) GetPageState() (string, error) {
	if sc.currentURL == "" {
		return "Браузер еще не открыт. Используйте команду 'navigate' для открытия URL.", nil
	}

	state := fmt.Sprintf(`Текущая страница: %s

 Состояние (симуляция в безопасном режиме):

В безопасном режиме AI агент может:
1. Анализировать команды пользователя
2. Создавать планы действий
3. Открывать URL в системном браузере

Для полной автоматизации требуется:
• Отключение антивируса для Rod/Playwright
• Или использование Selenium WebDriver
• Или работа на Linux/Mac

На странице обычно присутствуют:
- Заголовок страницы
- Навигационное меню
- Основной контент
- Формы ввода
- Кнопки действий

🎯 AI агент будет анализировать эту страницу и планировать следующие действия.`, sc.currentURL)

	return state, nil
}

func (sc *SafeController) GetDOM() (string, error) {
	dom := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Симуляция DOM - Безопасный режим</title>
</head>
<body>
    <header>
        <nav>
            <a href="/">Главная</a>
            <a href="/about">О нас</a>
            <a href="/contact">Контакты</a>
        </nav>
    </header>
    
    <main>
        <h1>AI Browser Agent - Симуляция</h1>
        <p>Это симулированный DOM в безопасном режиме.</p>
        
        <section id="search">
            <h2>Поиск</h2>
            <input type="text" id="search-input" placeholder="Введите запрос...">
            <button id="search-button">Найти</button>
        </section>
        
        <section id="content">
            <h2>Основной контент</h2>
            <p>В реальном режиме здесь был бы контент страницы %s</p>
            <div class="article">
                <h3>Статья 1</h3>
                <p>Содержание статьи...</p>
                <a href="/article/1" class="read-more">Читать далее</a>
            </div>
            <div class="article">
                <h3>Статья 2</h3>
                <p>Содержание статьи...</p>
                <a href="/article/2" class="read-more">Читать далее</a>
            </div>
        </section>
        
        <section id="forms">
            <h2>Формы</h2>
            <form id="login-form">
                <input type="email" placeholder="Email" name="email">
                <input type="password" placeholder="Пароль" name="password">
                <button type="submit">Войти</button>
            </form>
            
            <form id="contact-form">
                <input type="text" placeholder="Имя" name="name">
                <textarea placeholder="Сообщение" name="message"></textarea>
                <button type="submit">Отправить</button>
            </form>
        </section>
    </main>
    
    <footer>
        <p>&copy; 2024 AI Browser Agent</p>
    </footer>
</body>
</html>`, sc.currentURL)

	return dom, nil
}

func (sc *SafeController) Click(selector string) error {
	log.Printf(" Симуляция клика: %s", selector)
	log.Printf(" В реальном режиме был бы клик по элементу с селектором: %s", selector)
	return nil
}

// func (sc *SafeController) Click(selector string) error {
// 	log.Printf(" Симуляция клика: %s", selector)
// 	log.Printf(" В реальном режиме был бы клик по элементу с селектором: %s", selector)

// 	// Симулируем задержку клика
// 	time.Sleep(500 * time.Millisecond)
// 	return nil
// }

func (sc *SafeController) Fill(selector, value string) error {
	log.Printf(" Симуляция ввода текста '%s' в элемент: %s", value, selector)
	return nil
}

// func (sc *SafeController) Fill(selector, value string) error {
// 	log.Printf(" Симуляция ввода: %s = %s", selector, value)
// 	log.Printf(" В реальном режиме в поле %s был бы введен текст: %s", selector, value)

// 	// Симулируем задержку ввода
// 	time.Sleep(300 * time.Millisecond)
// 	return nil
// }

func (sc *SafeController) Screenshot(path string) error {
	log.Printf(" Симуляция скриншота: %s", path)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

// func (sc *SafeController) Screenshot(path string) error {
// 	if path == "" {
// 		path = fmt.Sprintf("screenshot_%d.txt", time.Now().Unix())
// 	}

// 	log.Printf(" Симуляция скриншота: %s", path)

// 	content := fmt.Sprintf(`AI Browser Agent - Скриншот (симуляция)

// Дата: %s
// Время: %s
// Режим: Безопасный (без автоматизации браузера)
// Текущий URL: %s

//  Информация о выполнении:
// • AI агент работает в безопасном режиме
// • Все команды анализируются и планируются
// • URL открываются в системном браузере
// • Взаимодействия симулируются

//  Для получения реальных скриншотов:
// 1. Отключите антивирус для Rod/Playwright
// 2. Или используйте Selenium WebDriver
// 3. Или запустите на Linux/Mac

//  Следующие шаги развития:
// • Интеграция с реальным браузерным движком
// • Компьютерное зение для анализа страниц
// • Векторная память для контекста
// • Поддержка сложных многошаговых задач`,
// 		time.Now().Format("2006-01-02"),
// 		time.Now().Format("15:04:05"),
// 		sc.currentURL)

// 	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
// 		return fmt.Errorf("не удалось сохранить скриншот: %w", err)
// 	}

// 	log.Printf("✅ Текстовый скриншот сохранен: %s", path)
// 	return nil
// }

func (sc *SafeController) Close() {
	log.Println(" Закрываю контроллер браузера (safe mode)")
}
